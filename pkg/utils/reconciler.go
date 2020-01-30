package utils

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	appsodyv1beta1 "github.com/appsody/appsody-operator/pkg/apis/appsody/v1beta1"
	"github.com/appsody/appsody-operator/pkg/common"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	servicebindingv1alpha1 "github.com/redhat-developer/service-binding-operator/pkg/apis/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

// ReconcilerBase base reconciler with some common behaviour
type ReconcilerBase struct {
	client     client.Client
	scheme     *runtime.Scheme
	recorder   record.EventRecorder
	restConfig *rest.Config
	discovery  discovery.DiscoveryInterface
}

//NewReconcilerBase creates a new ReconcilerBase
func NewReconcilerBase(client client.Client, scheme *runtime.Scheme, restConfig *rest.Config, recorder record.EventRecorder) ReconcilerBase {
	return ReconcilerBase{
		client:     client,
		scheme:     scheme,
		recorder:   recorder,
		restConfig: restConfig,
	}
}

// GetClient returns client
func (r *ReconcilerBase) GetClient() client.Client {
	return r.client
}

// GetRecorder returns the underlying recorder
func (r *ReconcilerBase) GetRecorder() record.EventRecorder {
	return r.recorder
}

// GetDiscoveryClient ...
func (r *ReconcilerBase) GetDiscoveryClient() (discovery.DiscoveryInterface, error) {
	if r.discovery == nil {
		var err error
		r.discovery, err = discovery.NewDiscoveryClientForConfig(r.restConfig)
		return r.discovery, err
	}

	return r.discovery, nil
}

// SetDiscoveryClient ...
func (r *ReconcilerBase) SetDiscoveryClient(discovery discovery.DiscoveryInterface) {
	r.discovery = discovery
}

var log = logf.Log.WithName("utils")

// CreateOrUpdate ...
func (r *ReconcilerBase) CreateOrUpdate(obj metav1.Object, owner metav1.Object, reconcile func() error) error {

	if owner != nil {
		controllerutil.SetControllerReference(owner, obj, r.scheme)
	}
	runtimeObj, ok := obj.(runtime.Object)
	if !ok {
		err := fmt.Errorf("%T is not a runtime.Object", obj)
		log.Error(err, "Failed to convert into runtime.Object")
		return err
	}
	result, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), runtimeObj, reconcile)
	if err != nil {
		return err
	}

	var gvk schema.GroupVersionKind
	gvk, err = apiutil.GVKForObject(runtimeObj, r.scheme)
	if err == nil {
		log.Info("Reconciled", "Kind", gvk.Kind, "Name", obj.GetName(), "Status", result)
	}

	return err
}

// DeleteResource deletes kubernetes resource
func (r *ReconcilerBase) DeleteResource(obj runtime.Object) error {
	err := r.client.Delete(context.TODO(), obj)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "Unable to delete object ", "object", obj)
			return err
		}
		return nil
	}

	metaObj, ok := obj.(metav1.Object)
	if !ok {
		err := fmt.Errorf("%T is not a metav1.Object", obj)
		log.Error(err, "Failed to convert into metav1.Object")
		return err
	}

	var gvk schema.GroupVersionKind
	gvk, err = apiutil.GVKForObject(obj, r.scheme)
	if err == nil {
		log.Info("Reconciled", "Kind", gvk.Kind, "Name", metaObj.GetName(), "Status", "deleted")
	}
	return nil
}

// DeleteResources ...
func (r *ReconcilerBase) DeleteResources(resources []runtime.Object) error {
	for i := range resources {
		err := r.DeleteResource(resources[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// GetAppsodyOpConfigMap ...
func (r *ReconcilerBase) GetAppsodyOpConfigMap(name string, ns string) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	err := r.GetClient().Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ns}, configMap)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}

// ManageError ...
func (r *ReconcilerBase) ManageError(issue error, conditionType common.StatusConditionType, ba common.BaseApplication) (reconcile.Result, error) {
	s := ba.GetStatus()
	rObj := ba.(runtime.Object)
	mObj := ba.(metav1.Object)
	logger := log.WithValues("ba.Namespace", mObj.GetNamespace(), "ba.Name", mObj.GetName())
	logger.Error(issue, "ManageError", "Condition", conditionType, "ba", ba)
	r.GetRecorder().Event(rObj, "Warning", "ProcessingError", issue.Error())

	oldCondition := s.GetCondition(conditionType)
	if oldCondition == nil {
		oldCondition = &appsodyv1beta1.StatusCondition{LastUpdateTime: metav1.Time{}}
	}

	lastUpdate := oldCondition.GetLastUpdateTime().Time
	lastStatus := oldCondition.GetStatus()

	// Keep the old `LastTransitionTime` when status has not changed
	nowTime := metav1.Now()
	transitionTime := oldCondition.GetLastTransitionTime()
	if lastStatus == corev1.ConditionTrue {
		transitionTime = &nowTime
	}

	newCondition := s.NewCondition()
	newCondition.SetLastTransitionTime(transitionTime)
	newCondition.SetLastUpdateTime(nowTime)
	newCondition.SetReason(string(apierrors.ReasonForError(issue)))
	newCondition.SetType(conditionType)
	newCondition.SetMessage(issue.Error())
	newCondition.SetStatus(corev1.ConditionFalse)

	s.SetCondition(newCondition)

	err := r.UpdateStatus(rObj)
	if err != nil {
		logger.Error(err, "Unable to update status")
		return reconcile.Result{
			RequeueAfter: time.Second,
			Requeue:      true,
		}, nil
	}

	// StatusReasonInvalid means the requested create or update operation cannot be
	// completed due to invalid data provided as part of the request. Don't retry.
	if apierrors.IsInvalid(issue) {
		return reconcile.Result{}, nil
	}

	var retryInterval time.Duration
	if lastUpdate.IsZero() || lastStatus == corev1.ConditionTrue {
		retryInterval = time.Second
	} else {
		retryInterval = newCondition.GetLastUpdateTime().Sub(lastUpdate).Round(time.Second)
	}

	return reconcile.Result{
		RequeueAfter: time.Duration(math.Min(float64(retryInterval.Nanoseconds()*2), float64(time.Hour.Nanoseconds()*6))),
		Requeue:      true,
	}, nil
}

// ManageSuccess ...
func (r *ReconcilerBase) ManageSuccess(conditionType common.StatusConditionType, ba common.BaseApplication) (reconcile.Result, error) {
	s := ba.GetStatus()
	oldCondition := s.GetCondition(conditionType)
	if oldCondition == nil {
		oldCondition = &appsodyv1beta1.StatusCondition{LastUpdateTime: metav1.Time{}}
	}

	// Keep the old `LastTransitionTime` when status has not changed
	nowTime := metav1.Now()
	transitionTime := oldCondition.GetLastTransitionTime()
	if oldCondition.GetStatus() == corev1.ConditionFalse {
		transitionTime = &nowTime
	}

	statusCondition := s.NewCondition()
	statusCondition.SetLastTransitionTime(transitionTime)
	statusCondition.SetLastUpdateTime(nowTime)
	statusCondition.SetReason("")
	statusCondition.SetMessage("")
	statusCondition.SetStatus(corev1.ConditionTrue)
	statusCondition.SetType(conditionType)

	s.SetCondition(statusCondition)
	err := r.UpdateStatus(ba.(runtime.Object))
	if err != nil {
		log.Error(err, "Unable to update status")
		return reconcile.Result{
			RequeueAfter: time.Second,
			Requeue:      true,
		}, nil
	}
	return reconcile.Result{}, nil
}

// IsGroupVersionSupported ...
func (r *ReconcilerBase) IsGroupVersionSupported(groupVersion string) (bool, error) {
	cli, err := r.GetDiscoveryClient()
	if err != nil {
		log.Error(err, "Failed to return a discovery client for the current reconciler")
		return false, err
	}

	_, err = cli.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// UpdateStatus updates the fields corresponding to the status subresource for the object
func (r *ReconcilerBase) UpdateStatus(obj runtime.Object) error {
	return r.GetClient().Status().Update(context.Background(), obj)
}

// SyncSecretAcrossNamespace syncs up the secret data across a namespace
func (r *ReconcilerBase) SyncSecretAcrossNamespace(fromSecret *corev1.Secret, namespace string) error {
	toSecret := &corev1.Secret{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: fromSecret.Name, Namespace: namespace}, toSecret)
	if err != nil {
		return err
	}
	toSecret.Data = fromSecret.Data
	return r.client.Update(context.TODO(), toSecret)
}

// AsOwner returns an owner reference object based on the input object. Also can set controller field on the owner ref.
func (r *ReconcilerBase) AsOwner(rObj runtime.Object, controller bool) (metav1.OwnerReference, error) {
	mObj, ok := rObj.(metav1.Object)
	if !ok {
		err := errors.Errorf("%T is not a metav1.Object", rObj)
		log.Error(err, "failed to convert into metav1.Object")
		return metav1.OwnerReference{}, err
	}

	gvk, err := apiutil.GVKForObject(rObj, r.scheme)
	if err != nil {
		log.Error(err, "failed to get GroupVersionKind associated with the runtime.Object", mObj)
		return metav1.OwnerReference{}, err
	}

	return metav1.OwnerReference{
		APIVersion: gvk.Group + "/" + gvk.Version,
		Kind:       gvk.Kind,
		Name:       mObj.GetName(),
		UID:        mObj.GetUID(),
		Controller: &controller,
	}, nil
}

// GetServiceBindingCreds returns a map containing username/password string values based on 'cr.spec.service.provides.auth'
func (r *ReconcilerBase) GetServiceBindingCreds(ba common.BaseApplication) (map[string]string, error) {
	if ba.GetService() == nil || ba.GetService().GetProvides() == nil || ba.GetService().GetProvides().GetAuth() == nil {
		return nil, errors.Errorf("auth is not set on the object %s", ba)
	}
	metaObj := ba.(metav1.Object)
	authMap := map[string]string{}

	auth := ba.GetService().GetProvides().GetAuth()
	getCred := func(key string, getCredF func() corev1.SecretKeySelector) error {
		if getCredF() != (corev1.SecretKeySelector{}) {
			cred, err := getCredFromSecret(metaObj.GetNamespace(), getCredF(), key, r.client)
			if err != nil {
				return err
			}
			authMap[key] = cred
		}
		return nil
	}
	err := getCred("username", auth.GetUsername)
	err = getCred("password", auth.GetPassword)
	if err != nil {
		return nil, err
	}
	return authMap, nil
}

func getCredFromSecret(namespace string, sel corev1.SecretKeySelector, cred string, client client.Client) (string, error) {
	secret := &corev1.Secret{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: sel.Name, Namespace: namespace}, secret)
	if err != nil {
		return "", errors.Wrapf(err, "unable to fetch credential %q from secret %q", cred, sel.Name)
	}

	if s, ok := secret.Data[sel.Key]; ok {
		return string(s), nil
	}
	return "", errors.Errorf("unable to find credential %q in secret %q using key %q", cred, sel.Name, sel.Key)
}

// ReconcileProvides ...
func (r *ReconcilerBase) ReconcileProvides(ba common.BaseApplication) (_ reconcile.Result, err error) {
	mObj := ba.(metav1.Object)
	logger := log.WithValues("ba.Namespace", mObj.GetNamespace(), "ba.Name", mObj.GetName())

	secretName := BuildServiceBindingSecretName(ba, mObj.GetName(), mObj.GetNamespace(), common.ServiceBindingCategoryOpenAPI)
	if ba.GetService().GetProvides() != nil && ba.GetService().GetProvides().GetCategory() == common.ServiceBindingCategoryOpenAPI {
		var creds map[string]string
		if ba.GetService().GetProvides().GetAuth() != nil {
			if creds, err = r.GetServiceBindingCreds(ba); err != nil {
				r.ManageError(errors.Wrapf(err, "service binding dependency not satisfied"), common.StatusConditionTypeDependenciesSatisfied, ba)
				return r.ManageError(errors.New("failed to get authentication info"), common.StatusConditionTypeReconciled, ba)
			}
		}

		secretMeta := metav1.ObjectMeta{
			Name:      secretName,
			Namespace: mObj.GetNamespace(),
		}
		providerSecret := &corev1.Secret{ObjectMeta: secretMeta}
		err = r.CreateOrUpdate(providerSecret, mObj, func() error {
			CustomizeServieBindingSecret(providerSecret, creds, ba)
			return nil
		})
		if err != nil {
			logger.Error(err, "Failed to reconcile provider secret")
			return r.ManageError(err, common.StatusConditionTypeReconciled, ba)
		}

		r.ManageSuccess(common.StatusConditionTypeDependenciesSatisfied, ba)
	} else {
		providerSecret := &corev1.Secret{}
		err = r.GetClient().Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: mObj.GetNamespace()}, providerSecret)
		if err != nil {
			if kerrors.IsNotFound(err) {
				logger.V(4).Info(fmt.Sprintf("Unable to find secret %q in namespace %q", secretName, mObj.GetNamespace()))
			} else {
				return r.ManageError(err, common.StatusConditionTypeReconciled, ba)
			}
		} else {
			// Delete all copies of this secret in other namespaces
			if providerSecret.Annotations["service."+ba.GetGroupName()+"/copied-to-namespaces"] != "" {
				namespaces := strings.Split(providerSecret.Annotations["service."+ba.GetGroupName()+"/copied-to-namespaces"], ",")
				for _, ns := range namespaces {
					err = r.GetClient().Delete(context.TODO(), &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: ns}})
					if err != nil {
						if kerrors.IsNotFound(err) {
							logger.V(4).Info(fmt.Sprintf("unable to find secret %q in namespace %q", secretName, mObj.GetNamespace()))
						} else {
							return r.ManageError(err, common.StatusConditionTypeReconciled, ba)
						}
					}
				}
			}

			// Delete the secret itself
			err = r.GetClient().Delete(context.TODO(), &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: mObj.GetNamespace()}})
			if err != nil {
				return r.ManageError(err, common.StatusConditionTypeReconciled, ba)
			}
		}
	}

	return r.ManageSuccess(common.StatusConditionTypeDependenciesSatisfied, ba)
}

// ReconcileConsumes ...
func (r *ReconcilerBase) ReconcileConsumes(ba common.BaseApplication) (reconcile.Result, error) {
	rObj := ba.(runtime.Object)
	mObj := ba.(metav1.Object)
	sbrNamesToConsume := map[string]bool{}
	for _, con := range ba.GetService().GetConsumes() {
		if con.GetCategory() == common.ServiceBindingCategoryOpenAPI {
			secretName := BuildServiceBindingSecretName(ba, con.GetName(), con.GetNamespace(), con.GetCategory())
			existingSecret := &corev1.Secret{}
			err := r.GetClient().Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: con.GetNamespace()}, existingSecret)
			if err != nil {
				if kerrors.IsNotFound(err) {
					delErr := r.DeleteResource(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: mObj.GetNamespace()}})
					if delErr != nil && !kerrors.IsNotFound(delErr) {
						delErr = errors.Wrapf(delErr, "unable to delete orphaned secret %q from namespace %q", secretName, mObj.GetNamespace())
						err = errors.Wrapf(delErr, "unable to find service binding secret %q for service %q in namespace %q", secretName, con.GetName(), con.GetNamespace())
					} else {
						err = errors.Wrapf(err, "unable to find service binding secret %q for service %q in namespace %q", secretName, con.GetName(), con.GetNamespace())
					}
				}
				r.ManageError(errors.Wrapf(err, "service binding dependency not satisfied"), common.StatusConditionTypeDependenciesSatisfied, ba)
				return r.ManageError(errors.New("dependency not satisfied"), common.StatusConditionTypeReconciled, ba)
			}

			if existingSecret.Annotations == nil {
				existingSecret.Annotations = map[string]string{}
			}
			existingSecret.Annotations["service."+ba.GetGroupName()+"/copied-to-namespaces"] =
				AppendIfNotSubstring(mObj.GetNamespace(), existingSecret.Annotations["service."+ba.GetGroupName()+"/copied-to-namespaces"])
			err = r.GetClient().Update(context.TODO(), existingSecret)
			if err != nil {
				r.ManageError(errors.Wrapf(err, "failed to update service provider secret"), common.StatusConditionTypeDependenciesSatisfied, ba)
				return r.ManageError(err, common.StatusConditionTypeReconciled, ba)
			}

			copiedSecret := &corev1.Secret{}
			err = r.GetClient().Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: mObj.GetNamespace()}, copiedSecret)
			if kerrors.IsNotFound(err) {
				owner, _ := r.AsOwner(rObj, false)
				copiedSecret = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:            secretName,
						Namespace:       mObj.GetNamespace(),
						Labels:          existingSecret.Labels,
						OwnerReferences: []metav1.OwnerReference{owner},
						Annotations:     map[string]string{"service." + ba.GetGroupName() + "/consumed-by": mObj.GetName()},
					},
					Data: existingSecret.Data,
				}
				err = r.GetClient().Create(context.TODO(), copiedSecret)
			} else if err == nil {
				existingCopiedSecret := copiedSecret.DeepCopyObject()
				if copiedSecret.Annotations == nil {
					copiedSecret.Annotations = map[string]string{}
				}
				copiedSecret.Annotations["service."+ba.GetGroupName()+"/consumed-by"] = AppendIfNotSubstring(mObj.GetName(), copiedSecret.Annotations["service."+ba.GetGroupName()+"/consumed-by"])
				copiedSecret.Data = existingSecret.Data
				// Skip setting the owner on the copiedSecret if the consumer and provider are in the same namespace
				// This is because we want the secret to be deleted if the provider is deleted
				if con.GetNamespace() != copiedSecret.Namespace {
					owner, _ := r.AsOwner(rObj, false)
					EnsureOwnerRef(copiedSecret, owner)
				}
				if !reflect.DeepEqual(existingCopiedSecret, copiedSecret) {
					err = r.GetClient().Update(context.TODO(), copiedSecret)
				}
			}

			if err != nil {
				r.ManageError(errors.Wrapf(err, "failed to create or update secret for consumes"), common.StatusConditionTypeDependenciesSatisfied, ba)
				return r.ManageError(err, common.StatusConditionTypeReconciled, ba)
			}

			consumedServices := ba.GetStatus().GetConsumedServices()
			if consumedServices == nil {
				consumedServices = common.ConsumedServices{}
			}
			if !ContainsString(consumedServices[common.ServiceBindingCategoryOpenAPI], secretName) {
				consumedServices[common.ServiceBindingCategoryOpenAPI] =
					append(consumedServices[common.ServiceBindingCategoryOpenAPI], secretName)
				ba.GetStatus().SetConsumedServices(consumedServices)
				err := r.UpdateStatus(rObj)
				if err != nil {
					r.ManageError(errors.Wrapf(err, "unable to update status with service binding secret information"), common.StatusConditionTypeDependenciesSatisfied, ba)
					return r.ManageError(err, common.StatusConditionTypeReconciled, ba)
				}
			}
		} else if con.GetCategory() == common.ServiceBindingCategoryResource {
			if ok, err := r.IsGroupVersionSupported(servicebindingv1alpha1.SchemeGroupVersion.String()); err != nil {
				return r.ManageError(errors.New("failed to check if Service Binding Request CRD is installed"), common.StatusConditionTypeReconciled, ba)
			} else if ok {
				if con.GetNamespace() == mObj.GetNamespace() {
					sbrName := BuildServiceBindingSecretName(ba, con.GetName(), con.GetNamespace(), con.GetCategory())
					sbrNamesToConsume[sbrName] = true
					sbr := &servicebindingv1alpha1.ServiceBindingRequest{
						ObjectMeta: metav1.ObjectMeta{
							Name:      sbrName,
							Namespace: mObj.GetNamespace(),
						}}
					err = r.CreateOrUpdate(sbr, mObj, func() error {
						CustomizeServieBindingRequest(sbr, con, ba)
						// TODO :: add a dummy deployment. MUST BE REMOVED ONCE SBO MAKES APP SELECTOR OPTIONAL
						sbr.Spec.ApplicationSelector = servicebindingv1alpha1.ApplicationSelector{
							Group:       "extensions",
							Version:     "v1beta1",
							Resource:    "deployments",
							ResourceRef: mObj.GetName(),
						}
						return nil
					})
					if err != nil {
						r.ManageError(errors.Wrapf(err, "service binding dependency not satisfied"), common.StatusConditionTypeDependenciesSatisfied, ba)
						return r.ManageError(errors.New("dependency not satisfied"), common.StatusConditionTypeReconciled, ba)
					}

					if sbr.Status.BindingStatus != "Success" {
						r.ManageError(errors.Errorf("service binding dependency not satisfied. statuc field for %q Service Binding Request's status (%q) is not \"Success\". ", sbr.Name, sbr.Status.BindingStatus), common.StatusConditionTypeDependenciesSatisfied, ba)
						return r.ManageError(errors.New("dependency not satisfied"), common.StatusConditionTypeReconciled, ba)
					}

					bindingSecret := &corev1.Secret{}
					err = r.GetClient().Get(context.TODO(), types.NamespacedName{Name: sbr.Status.Secret, Namespace: mObj.GetNamespace()}, bindingSecret)
					if err != nil {
						if kerrors.IsNotFound(err) {
							err = errors.Wrapf(err, "unable to find service binding secret %q for Service Binding Request resource in namespace %q", sbrName, mObj.GetNamespace())
						}
						r.ManageError(errors.Wrapf(err, "service binding dependency not satisfied"), common.StatusConditionTypeDependenciesSatisfied, ba)
						return r.ManageError(errors.New("dependency not satisfied"), common.StatusConditionTypeReconciled, ba)
					}

					consumedServices := ba.GetStatus().GetConsumedServices()
					if consumedServices == nil {
						consumedServices = common.ConsumedServices{}
					}
					if !ContainsString(consumedServices[common.ServiceBindingCategoryResource], sbr.Status.Secret) {
						consumedServices[common.ServiceBindingCategoryResource] =
							append(consumedServices[common.ServiceBindingCategoryResource], sbr.Status.Secret)
						ba.GetStatus().SetConsumedServices(consumedServices)
						err := r.UpdateStatus(rObj)
						if err != nil {
							r.ManageError(errors.Wrapf(err, "unable to update status with service binding secret information"), common.StatusConditionTypeDependenciesSatisfied, ba)
							return r.ManageError(err, common.StatusConditionTypeReconciled, ba)
						}
					}
				} else {
					// TODO :: handle when backing service is in another namespace
				}
			} else {
				r.ManageError(errors.Errorf("could not find ServiceBindingRequest CRD (%q) in the cluster. ensure Service Binding Operator is installed", servicebindingv1alpha1.SchemeGroupVersion.String()), common.StatusConditionTypeDependenciesSatisfied, ba)
				return r.ManageError(errors.New("dependency not satisfied"), common.StatusConditionTypeReconciled, ba)
			}
		}
	}

	// Cleaning up leftover SBR resources if not consumes anymore
	if ok, _ := r.IsGroupVersionSupported(servicebindingv1alpha1.SchemeGroupVersion.String()); ok {
		sbrList := &servicebindingv1alpha1.ServiceBindingRequestList{}
		err := r.GetClient().List(context.Background(), sbrList, client.InNamespace(mObj.GetNamespace()), client.MatchingLabels(ba.GetLabels()))
		if err != nil {
			return r.ManageError(errors.New("failed to look up Service Binding Requests to clean up"), common.StatusConditionTypeReconciled, ba)
		}
		for _, sbr := range sbrList.Items {
			if !sbrNamesToConsume[sbr.Name] {
				err = r.DeleteResource(&sbr)
				if err != nil {
					return r.ManageError(errors.New("failed to delete Service Binding Requests"), common.StatusConditionTypeReconciled, ba)
				}
			}
		}
	}
	return r.ManageSuccess(common.StatusConditionTypeDependenciesSatisfied, ba)
}

// IsOpenShift returns true if the operator is running on an OpenShift platform
func (r *ReconcilerBase) IsOpenShift() bool {
	isOpenShift, err := r.IsGroupVersionSupported(routev1.SchemeGroupVersion.String())
	if err != nil {
		return false
	}
	return isOpenShift
}
