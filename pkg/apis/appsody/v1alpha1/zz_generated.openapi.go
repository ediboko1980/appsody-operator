// +build !ignore_autogenerated

// Code generated by openapi-gen. DO NOT EDIT.

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplication":            schema_pkg_apis_appsody_v1alpha1_AppsodyApplication(ref),
		"github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationAutoScaling": schema_pkg_apis_appsody_v1alpha1_AppsodyApplicationAutoScaling(ref),
		"github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationService":     schema_pkg_apis_appsody_v1alpha1_AppsodyApplicationService(ref),
		"github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationSpec":        schema_pkg_apis_appsody_v1alpha1_AppsodyApplicationSpec(ref),
		"github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationStatus":      schema_pkg_apis_appsody_v1alpha1_AppsodyApplicationStatus(ref),
		"github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationStorage":     schema_pkg_apis_appsody_v1alpha1_AppsodyApplicationStorage(ref),
	}
}

func schema_pkg_apis_appsody_v1alpha1_AppsodyApplication(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "AppsodyApplication is the Schema for the appsodyapplications API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationSpec", "github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_appsody_v1alpha1_AppsodyApplicationAutoScaling(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "AppsodyApplicationAutoScaling ...",
				Properties: map[string]spec.Schema{
					"targetCPUUtilizationPercentage": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"minReplicas": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"maxReplicas": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
				},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_appsody_v1alpha1_AppsodyApplicationService(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "AppsodyApplicationService ...",
				Properties: map[string]spec.Schema{
					"type": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"port": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
				},
				Required: []string{"port"},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_appsody_v1alpha1_AppsodyApplicationSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "AppsodyApplicationSpec defines the desired state of AppsodyApplication",
				Properties: map[string]spec.Schema{
					"applicationImage": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"replicas": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"autoscaling": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationAutoScaling"),
						},
					},
					"pullPolicy": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"pullSecret": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"volumes": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/api/core/v1.Volume"),
									},
								},
							},
						},
					},
					"volumeMounts": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/api/core/v1.VolumeMount"),
									},
								},
							},
						},
					},
					"resourceConstraints": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/api/core/v1.ResourceRequirements"),
						},
					},
					"readinessProbe": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/api/core/v1.Probe"),
						},
					},
					"livenessProbe": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/api/core/v1.Probe"),
						},
					},
					"service": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationService"),
						},
					},
					"expose": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"envFrom": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/api/core/v1.EnvFromSource"),
									},
								},
							},
						},
					},
					"env": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/api/core/v1.EnvVar"),
									},
								},
							},
						},
					},
					"serviceAccountName": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"architecture": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"string"},
										Format: "",
									},
								},
							},
						},
					},
					"storage": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationStorage"),
						},
					},
					"createKnativeService": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"stack": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
				Required: []string{"applicationImage"},
			},
		},
		Dependencies: []string{
			"github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationAutoScaling", "github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationService", "github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationStorage", "k8s.io/api/core/v1.EnvFromSource", "k8s.io/api/core/v1.EnvVar", "k8s.io/api/core/v1.Probe", "k8s.io/api/core/v1.ResourceRequirements", "k8s.io/api/core/v1.Volume", "k8s.io/api/core/v1.VolumeMount"},
	}
}

func schema_pkg_apis_appsody_v1alpha1_AppsodyApplicationStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "AppsodyApplicationStatus defines the observed state of AppsodyApplication",
				Properties: map[string]spec.Schema{
					"conditions": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationStatusCondition"),
									},
								},
							},
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/appsody-operator/pkg/apis/appsody/v1alpha1.AppsodyApplicationStatusCondition"},
	}
}

func schema_pkg_apis_appsody_v1alpha1_AppsodyApplicationStorage(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "AppsodyApplicationStorage ...",
				Properties: map[string]spec.Schema{
					"size": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"mountPath": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"volumeClaimTemplate": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/api/core/v1.PersistentVolumeClaim"),
						},
					},
				},
				Required: []string{"mountPath"},
			},
		},
		Dependencies: []string{
			"k8s.io/api/core/v1.PersistentVolumeClaim"},
	}
}
