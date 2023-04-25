// Copyright © 2021 Cisco Systems, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nodeagent

import (
	"fmt"

	"emperror.dev/errors"
	"github.com/cisco-open/operator-tools/pkg/merge"
	"github.com/cisco-open/operator-tools/pkg/reconciler"
	"github.com/cisco-open/operator-tools/pkg/typeoverride"
	util "github.com/cisco-open/operator-tools/pkg/utils"
	"github.com/go-logr/logr"
	"github.com/kube-logging/logging-operator/pkg/resources"
	"github.com/kube-logging/logging-operator/pkg/resources/fluentddataprovider"
	"github.com/kube-logging/logging-operator/pkg/sdk/logging/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	serviceAccountNameFluentbit     = "fluentbit"
	clusterRoleBindingNameFluentbit = "fluentbit"
	clusterRoleNameFluentbit        = "fluentbit"
	SecretConfigNameFluentbit       = "fluentbit"
	DaemonSetNameFluentbit          = "fluentbit"
	PodSecurityPolicyNameFluentbit  = "fluentbit"
	ServiceNameFluentbit            = "fluentbit"
	containerNameFluentbit          = "fluent-bit"
)
const (
	serviceAccountNameSyslogNG                = "syslog-ng"
	secretConfigNameSyslogNG                  = "syslog-ng"
	DaemonSetNameSyslogNG                     = "syslog-ng"
	PodSecurityPolicyNameSyslogNG             = "syslog-ng"
	serviceNameSyslogNG                       = "syslog-ng"
	imageImageRepositorySyslogNG              = "ghcr.io/axoflow/syslog-ng"
	ServicePortSyslogNG                       = 601
	configSecretNameSyslogNG                  = "syslog-ng"
	configKeySyslogNG                         = "syslog-ng.conf"
	statefulSetNameSyslogNG                   = "syslog-ng"
	outputSecretNameSyslogNG                  = "syslog-ng-output"
	OutputSecretPathSyslogNG                  = "/etc/syslog-ng/secret"
	bufferPathSyslogNG                        = "/buffers"
	roleBindingNameSyslogNG                   = "syslog-ng"
	roleNameSyslogNG                          = "syslog-ng"
	clusterRoleBindingNameSyslogNG            = "syslog-ng"
	clusterRoleNameSyslogNG                   = "syslog-ng"
	containerNameSyslogNG                     = "syslog-ng"
	defaultBufferVolumeMetricsPortSyslogNG    = 9200
	imageRepositorySyslogNG                   = "ghcr.io/axoflow/syslog-ng"
	imageTagSyslogNG                          = "4.1.1"
	prometheusExporterImageRepositorySyslogNG = "ghcr.io/kube-logging/syslog-ng-exporter"
	prometheusExporterImageTagSyslogNG        = "v0.0.14"
	bufferVolumeImageRepositorySyslogNG       = "ghcr.io/kube-logging/node-exporter"
	bufferVolumeImageTagSyslogNG              = "v0.2.0"
	configReloaderImageRepositorySyslogNG     = "ghcr.io/kube-logging/syslogng-reload"
	configReloaderImageTagSyslogNG            = "v1.0.1"
	socketVolumeNameSyslogNG                  = "socket"
	socketPathSyslogNG                        = "/tmp/syslog-ng/syslog-ng.ctl"
	configDirSyslogNG                         = "/etc/syslog-ng/config"
	configVolumeNameSyslogNG                  = "config"
	tlsVolumeNameSyslogNG                     = "tls"
	metricsPortNumberSyslogNG                 = 9577
	metricsPortNameSyslogNG                   = "exporter"
)

func NodeAgentFluentbitDefaults(userDefined v1beta1.NodeAgentConfig) (*v1beta1.NodeAgentConfig, error) {
	programDefault := &v1beta1.NodeAgentConfig{
		FluentbitSpec: &v1beta1.NodeAgentFluentbit{
			DaemonSetOverrides: &typeoverride.DaemonSet{
				Spec: typeoverride.DaemonSetSpec{
					Template: typeoverride.PodTemplateSpec{
						ObjectMeta: typeoverride.ObjectMeta{
							Annotations: map[string]string{},
						},
						Spec: typeoverride.PodSpec{
							Containers: []v1.Container{
								{
									Name:            containerNameFluentbit,
									Image:           "fluent/fluent-bit:1.9.10",
									Command:         []string{"/fluent-bit/bin/fluent-bit", "-c", "/fluent-bit/conf_operator/fluent-bit.conf"},
									ImagePullPolicy: v1.PullIfNotPresent,
									Resources: v1.ResourceRequirements{
										Limits: v1.ResourceList{
											v1.ResourceMemory: resource.MustParse("100M"),
											v1.ResourceCPU:    resource.MustParse("200m"),
										},
										Requests: v1.ResourceList{
											v1.ResourceMemory: resource.MustParse("50M"),
											v1.ResourceCPU:    resource.MustParse("100m"),
										},
									},
								},
							},
						},
					},
				},
			},
			Flush:         1,
			Grace:         5,
			LogLevel:      "info",
			CoroStackSize: 24576,
			InputTail: v1beta1.InputTail{
				Path:            "/var/log/containers/*.log",
				RefreshInterval: "5",
				SkipLongLines:   "On",
				DB:              util.StringPointer("/tail-db/tail-containers-state.db"),
				MemBufLimit:     "5MB",
				Tag:             "kubernetes.*",
			},
			Security: &v1beta1.Security{
				RoleBasedAccessControlCreate: util.BoolPointer(true),
				SecurityContext:              &v1.SecurityContext{},
				PodSecurityContext:           &v1.PodSecurityContext{},
			},
			ContainersPath: "/var/lib/docker/containers",
			VarLogsPath:    "/var/log",
			BufferStorage: v1beta1.BufferStorage{
				StoragePath: "/buffers",
			},

			ForwardOptions: &v1beta1.ForwardOptions{
				RetryLimit: "False",
			},
		},
	}
	if userDefined.FluentbitSpec == nil {
		userDefined.FluentbitSpec = &v1beta1.NodeAgentFluentbit{}
	}

	if userDefined.FluentbitSpec.FilterAws != nil {
		programDefault.FluentbitSpec.FilterAws = &v1beta1.FilterAws{
			ImdsVersion:     "v2",
			AZ:              util.BoolPointer(true),
			Ec2InstanceID:   util.BoolPointer(true),
			Ec2InstanceType: util.BoolPointer(false),
			PrivateIP:       util.BoolPointer(false),
			AmiID:           util.BoolPointer(false),
			AccountID:       util.BoolPointer(false),
			Hostname:        util.BoolPointer(false),
			VpcID:           util.BoolPointer(false),
			Match:           "*",
		}

		err := merge.Merge(programDefault.FluentbitSpec.FilterAws, userDefined.FluentbitSpec.FilterAws)
		if err != nil {
			return nil, err
		}

	}
	if userDefined.FluentbitSpec.LivenessDefaultCheck == nil || *userDefined.FluentbitSpec.LivenessDefaultCheck {
		if userDefined.Profile != "windows" {
			programDefault.FluentbitSpec.Metrics = &v1beta1.Metrics{
				Port: 2020,
				Path: "/",
			}
		}
	}

	if userDefined.FluentbitSpec.Metrics != nil {

		programDefault.FluentbitSpec.Metrics = &v1beta1.Metrics{
			Interval: "15s",
			Timeout:  "5s",
			Port:     2020,
			Path:     "/api/v1/metrics/prometheus",
		}
		err := merge.Merge(programDefault.FluentbitSpec.Metrics, userDefined.FluentbitSpec.Metrics)
		if err != nil {
			return nil, err
		}
	}
	if programDefault.FluentbitSpec.Metrics != nil && userDefined.FluentbitSpec.Metrics != nil && userDefined.FluentbitSpec.Metrics.PrometheusAnnotations {
		defaultPrometheusAnnotations := &typeoverride.ObjectMeta{
			Annotations: map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/path":   programDefault.FluentbitSpec.Metrics.Path,
				"prometheus.io/port":   fmt.Sprintf("%d", programDefault.FluentbitSpec.Metrics.Port),
			},
		}
		err := merge.Merge(&(programDefault.FluentbitSpec.DaemonSetOverrides.Spec.Template.ObjectMeta), defaultPrometheusAnnotations)
		if err != nil {
			return nil, err
		}
	}
	if programDefault.FluentbitSpec.Metrics != nil {
		defaultLivenessProbe := &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: programDefault.FluentbitSpec.Metrics.Path,
					Port: intstr.IntOrString{
						IntVal: programDefault.FluentbitSpec.Metrics.Port,
					},
				}},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      0,
			PeriodSeconds:       10,
			SuccessThreshold:    0,
			FailureThreshold:    3,
		}
		if programDefault.FluentbitSpec.DaemonSetOverrides.Spec.Template.Spec.Containers[0].LivenessProbe == nil {
			programDefault.FluentbitSpec.DaemonSetOverrides.Spec.Template.Spec.Containers[0].LivenessProbe = &v1.Probe{}
		}

		err := merge.Merge(programDefault.FluentbitSpec.DaemonSetOverrides.Spec.Template.Spec.Containers[0].LivenessProbe, defaultLivenessProbe)
		if err != nil {
			return nil, err
		}
	}

	return programDefault, nil
}

func NodeAgentSyslogNGDefaults(userDefined v1beta1.NodeAgentConfig) (*v1beta1.NodeAgentConfig, error) {
	programDefault := &v1beta1.NodeAgentConfig{
		SyslogNGSpec: &v1beta1.NodeAgentSyslogNG{
			DaemonSetOverrides: &typeoverride.DaemonSet{
				Spec: typeoverride.DaemonSetSpec{
					Template: typeoverride.PodTemplateSpec{
						ObjectMeta: typeoverride.ObjectMeta{
							Annotations: map[string]string{},
						},
						Spec: typeoverride.PodSpec{
							Containers: []v1.Container{
								{
									Name:  containerNameSyslogNG,
									Image: v1beta1.RepositoryWithTag(imageRepositorySyslogNG, imageTagSyslogNG),
									Args: []string{
										"--cfgfile=" + configDirSyslogNG + "/" + configKeySyslogNG,
										"--control=" + socketPathSyslogNG,
										"--no-caps",
										"-Fe",
									},
									ImagePullPolicy: v1.PullIfNotPresent,
									Resources: v1.ResourceRequirements{
										Limits: v1.ResourceList{
											v1.ResourceMemory: resource.MustParse("400M"),
											v1.ResourceCPU:    resource.MustParse("1000m"),
										},
										Requests: v1.ResourceList{
											v1.ResourceMemory: resource.MustParse("100M"),
											v1.ResourceCPU:    resource.MustParse("500m"),
										},
									},
								},
							},
						},
					},
				},
			},
			Security: &v1beta1.Security{
				RoleBasedAccessControlCreate: util.BoolPointer(true),
				SecurityContext:              &v1.SecurityContext{},
				PodSecurityContext:           &v1.PodSecurityContext{},
			},
			ContainersPath: "/var/lib/docker/containers",
			VarLogsPath:    "/var/log",
			BufferStorage: v1beta1.BufferStorage{
				StoragePath: "/buffers",
			},
		},
	}
	if userDefined.SyslogNGSpec == nil {
		userDefined.SyslogNGSpec = &v1beta1.NodeAgentSyslogNG{}
	}

	if userDefined.SyslogNGSpec.Metrics != nil {
		// TODO implement the same as implemented in the aggregator
		if userDefined.SyslogNGSpec.Metrics.PrometheusAnnotations {
			defaultPrometheusAnnotations := &typeoverride.ObjectMeta{
				Annotations: map[string]string{
					"prometheus.io/scrape": "true",
					"prometheus.io/path":   programDefault.SyslogNGSpec.Metrics.Path,
					"prometheus.io/port":   fmt.Sprintf("%d", programDefault.SyslogNGSpec.Metrics.Port),
				},
			}
			err := merge.Merge(&(programDefault.SyslogNGSpec.DaemonSetOverrides.Spec.Template.ObjectMeta), defaultPrometheusAnnotations)
			if err != nil {
				return nil, err
			}
		}
		defaultLivenessProbe := &v1.Probe{
			ProbeHandler: v1.ProbeHandler{
				HTTPGet: &v1.HTTPGetAction{
					Path: programDefault.SyslogNGSpec.Metrics.Path,
					Port: intstr.IntOrString{
						IntVal: programDefault.SyslogNGSpec.Metrics.Port,
					},
				}},
			InitialDelaySeconds: 10,
			TimeoutSeconds:      0,
			PeriodSeconds:       10,
			SuccessThreshold:    0,
			FailureThreshold:    3,
		}
		if programDefault.SyslogNGSpec.DaemonSetOverrides.Spec.Template.Spec.Containers[0].LivenessProbe == nil {
			programDefault.SyslogNGSpec.DaemonSetOverrides.Spec.Template.Spec.Containers[0].LivenessProbe = &v1.Probe{}
		}

		err := merge.Merge(programDefault.SyslogNGSpec.DaemonSetOverrides.Spec.Template.Spec.Containers[0].LivenessProbe, defaultLivenessProbe)
		if err != nil {
			return nil, err
		}
	}

	return programDefault, nil
}

var NodeAgentFluentbitWindowsDefaults = &v1beta1.NodeAgentConfig{
	FluentbitSpec: &v1beta1.NodeAgentFluentbit{
		FilterKubernetes: v1beta1.FilterKubernetes{
			KubeURL:       "https://kubernetes.default.svc:443",
			KubeCAFile:    "c:\\var\\run\\secrets\\kubernetes.io\\serviceaccount\\ca.crt",
			KubeTokenFile: "c:\\var\\run\\secrets\\kubernetes.io\\serviceaccount\\token",
			KubeTagPrefix: "kubernetes.C.var.log.containers.",
		},
		InputTail: v1beta1.InputTail{
			Path: "C:\\var\\log\\containers\\*.log",
		},
		ContainersPath: "C:\\ProgramData\\docker",
		VarLogsPath:    "C:\\var\\log",
		DaemonSetOverrides: &typeoverride.DaemonSet{
			Spec: typeoverride.DaemonSetSpec{
				Template: typeoverride.PodTemplateSpec{
					Spec: typeoverride.PodSpec{
						Containers: []v1.Container{
							{
								Name:    containerNameFluentbit,
								Image:   "rancher/fluent-bit:1.6.10-rc7",
								Command: []string{"fluent-bit", "-c", "fluent-bit\\conf_operator\\fluent-bit.conf"},
								Resources: v1.ResourceRequirements{
									Limits: v1.ResourceList{
										v1.ResourceMemory: resource.MustParse("200M"),
										v1.ResourceCPU:    resource.MustParse("200m"),
									},
									Requests: v1.ResourceList{
										v1.ResourceMemory: resource.MustParse("100M"),
										v1.ResourceCPU:    resource.MustParse("100m"),
									},
								},
							}},
						NodeSelector: map[string]string{
							"kubernetes.io/os": "windows",
						},
					}},
			}},
	},
}
var NodeAgentFluentbitLinuxDefaults = &v1beta1.NodeAgentConfig{
	FluentbitSpec: &v1beta1.NodeAgentFluentbit{},
}

func generateLoggingRefLabels(loggingRef string) map[string]string {
	return map[string]string{"app.kubernetes.io/managed-by": loggingRef}
}

func (n *nodeAgentInstance) getNodeAgentLabels() map[string]string {
	if n.nodeAgent.FluentbitSpec != nil {
		return util.MergeLabels(n.nodeAgent.Metadata.Labels, map[string]string{
			"app.kubernetes.io/name":     "fluentbit",
			"app.kubernetes.io/instance": n.name,
		}, generateLoggingRefLabels(n.logging.ObjectMeta.GetName()))

	} else if n.nodeAgent.SyslogNGSpec != nil {
		return util.MergeLabels(n.nodeAgent.Metadata.Labels, map[string]string{
			"app.kubernetes.io/name":     "syslog-ng",
			"app.kubernetes.io/instance": n.name,
		}, generateLoggingRefLabels(n.logging.ObjectMeta.GetName()))

	}
	return nil
}

func (n *nodeAgentInstance) getServiceAccount() string {
	if n.nodeAgent.FluentbitSpec != nil {
		if n.nodeAgent.FluentbitSpec.Security != nil && n.nodeAgent.FluentbitSpec.Security.ServiceAccount != "" {
			return n.nodeAgent.FluentbitSpec.Security.ServiceAccount
		}
		return n.QualifiedName(serviceAccountNameFluentbit)
	}
	if n.nodeAgent.SyslogNGSpec != nil {
		if n.nodeAgent.SyslogNGSpec.Security != nil && n.nodeAgent.SyslogNGSpec.Security.ServiceAccount != "" {
			return n.nodeAgent.SyslogNGSpec.Security.ServiceAccount
		}
		return n.QualifiedName(serviceAccountNameSyslogNG)
	}
	return ""
}

//	type DesiredObject struct {
//		Object runtime.Object
//		State  reconciler.DesiredState
//	}
//
// Reconciler holds info what resource to reconcile
type Reconciler struct {
	Logging *v1beta1.Logging
	*reconciler.GenericResourceReconciler
	configs             map[string][]byte
	agents              map[string]v1beta1.NodeAgentConfig
	fluentdDataProvider fluentddataprovider.FluentdDataProvider
}

// New creates a new NodeAgent reconciler
func New(client client.Client, logger logr.Logger, logging *v1beta1.Logging, agents map[string]v1beta1.NodeAgentConfig, opts reconciler.ReconcilerOpts, fluentdDataProvider fluentddataprovider.FluentdDataProvider) *Reconciler {
	return &Reconciler{
		Logging:                   logging,
		GenericResourceReconciler: reconciler.NewGenericReconciler(client, logger, opts),
		agents:                    agents,
		fluentdDataProvider:       fluentdDataProvider,
	}
}

type nodeAgentInstance struct {
	name                string
	nodeAgent           *v1beta1.NodeAgentConfig
	reconciler          *reconciler.GenericResourceReconciler
	logging             *v1beta1.Logging
	configs             map[string][]byte
	fluentdDataProvider fluentddataprovider.FluentdDataProvider
}

// Reconcile reconciles the InlineNodeAgent resource
func (r *Reconciler) Reconcile() (*reconcile.Result, error) {
	combinedResult := reconciler.CombinedResult{}
	for name, userDefinedAgent := range r.agents {
		result, err := r.processAgent(name, userDefinedAgent)
		combinedResult.Combine(result, err)
	}
	return &combinedResult.Result, combinedResult.Err
}

func (r *Reconciler) processAgent(name string, userDefinedAgent v1beta1.NodeAgentConfig) (*reconcile.Result, error) {
	var instance nodeAgentInstance
	var nodeAgentConfig *v1beta1.NodeAgentConfig
	var err error

	if userDefinedAgent.FluentbitSpec != nil && userDefinedAgent.SyslogNGSpec != nil {
		return nil, errors.New("only one agent implementation can be specified for a single nodeAgent")
	}

	if userDefinedAgent.FluentbitSpec != nil {
		if nodeAgentConfig, err = NodeAgentFluentbitDefaults(userDefinedAgent); err != nil {
			return nil, err
		}
		switch userDefinedAgent.Profile {
		case "windows":
			if err := merge.Merge(nodeAgentConfig, NodeAgentFluentbitWindowsDefaults); err != nil {
				return nil, err
			}
			// Overwrite Kubernetes endpoint with a ClusterDomain templated value.
			nodeAgentConfig.FluentbitSpec.FilterKubernetes.KubeURL = fmt.Sprintf("https://kubernetes.default.svc%s:443", r.Logging.ClusterDomainAsSuffix())
		default:
			if err := merge.Merge(nodeAgentConfig, NodeAgentFluentbitLinuxDefaults); err != nil {
				return nil, err
			}
		}
	}

	if userDefinedAgent.SyslogNGSpec != nil {
		if nodeAgentConfig, err = NodeAgentSyslogNGDefaults(userDefinedAgent); err != nil {
			return nil, err
		}
	}

	err = merge.Merge(nodeAgentConfig, &userDefinedAgent)
	if err != nil {
		return nil, err
	}

	instance = nodeAgentInstance{
		name:                name,
		nodeAgent:           nodeAgentConfig,
		reconciler:          r.GenericResourceReconciler,
		logging:             r.Logging,
		fluentdDataProvider: r.fluentdDataProvider,
	}

	return instance.Reconcile()
}

// Reconcile reconciles the nodeAgentInstance resource
func (n *nodeAgentInstance) Reconcile() (*reconcile.Result, error) {
	for _, factory := range []resources.Resource{
		n.serviceAccount,
		n.clusterRole,
		n.clusterRoleBinding,
		n.clusterPodSecurityPolicy,
		n.pspClusterRole,
		n.pspClusterRoleBinding,
		n.configSecret,
		n.daemonSet,
		n.serviceMetrics,
		n.monitorServiceMetrics,
	} {
		o, state, err := factory()
		if err != nil {
			return nil, errors.WrapIf(err, "failed to create desired object")
		}
		if o == nil {
			return nil, errors.Errorf("Reconcile error! Resource %#v returns with nil object", factory)
		}
		result, err := n.reconciler.ReconcileResource(o, state)
		if err != nil {
			return nil, errors.WrapWithDetails(err,
				"failed to reconcile resource", "resource", o.GetObjectKind().GroupVersionKind())
		}
		if result != nil {
			return result, nil
		}
	}

	return nil, nil
}

func RegisterWatches(builder *builder.Builder) *builder.Builder {
	return builder.
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&rbacv1.ClusterRole{}).
		Owns(&rbacv1.ClusterRoleBinding{}).
		Owns(&corev1.ServiceAccount{})
}

// nodeAgent QualifiedName
func (n *nodeAgentInstance) QualifiedName(name string) string {
	return fmt.Sprintf("%s-%s-%s", n.logging.Name, n.name, name)
}

// nodeAgent AggregatorQualifiedName
func (n *nodeAgentInstance) AggregatorQualifiedName(name string) string {
	return fmt.Sprintf("%s-%s", n.logging.Name, name)
}
