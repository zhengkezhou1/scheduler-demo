package webhook

import (
	"context"
	"encoding/json"
	"fmt"

	"math"

	"github.com/scheduler-demo/pkg/kube"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"
)

func admitPods(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.Info("admitting pods")
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	if ar.Request.Resource != podResource {
		err := fmt.Errorf("expect resource to be %s", podResource)
		klog.Error(err)
		return ToV1AdmissionResponse(err)
	}

	raw := ar.Request.Object.Raw
	pod := corev1.Pod{}
	deserializer := serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
		klog.Error(err)
		return ToV1AdmissionResponse(err)
	}
	reviewResponse := v1.AdmissionResponse{}
	reviewResponse.Allowed = true

	patches := []map[string]any{
		{
			"op":   "add",
			"path": "/spec/affinity",
			"value": corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Weight: 100,
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "node.kubernetes.io/capacity",
										Operator: corev1.NodeSelectorOpIn,
										// TODO: it's ok for no status app, don't do that on status app.
										Values: []string{"spot"},
									},
								},
							},
						},
					},
				},
			},
		},
		topologySpreadConstraintsPatch(&pod),
	}

	patchBytes, err := json.Marshal(patches)
	if err != nil {
		klog.Error(err)
		return ToV1AdmissionResponse(err)
	}

	pt := v1.PatchTypeJSONPatch
	reviewResponse.PatchType = &pt
	reviewResponse.Patch = patchBytes
	return &reviewResponse
}

func topologySpreadConstraintsPatch(pod *corev1.Pod) map[string]any {
	return map[string]any{
		"op":   "add",
		"path": "/spec/topologySpreadConstraints",
		"value": []corev1.TopologySpreadConstraint{
			{
				MaxSkew:           calculateMaxSkewForPod(pod),
				TopologyKey:       "node.kubernetes.io/capacity",
				WhenUnsatisfiable: corev1.DoNotSchedule,
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: getSafeLabels(pod.Labels),
				},
			},
		},
	}
}

// calculateMaxSkewForPod dynamically adjusts based on pod replica count and actual node counts (spot, on-demand) in the cluster.
func calculateMaxSkewForPod(pod *corev1.Pod) int32 {
	schedulerNodes, err := getKubeNodeClient().ListSchedulerNodes()
	if err != nil || schedulerNodes == nil {
		klog.Error(err)
		return 1
	}

	spotNodeCount := len(schedulerNodes.SpotNodes)
	onDemandNodeCount := len(schedulerNodes.OnDemandNodes)

	random := int32(math.Abs(float64(spotNodeCount - onDemandNodeCount)))

	// Get Deployment replica count through OwnerReference
	replicas := getDeploymentReplicas(pod)
	if replicas == 1 {
		return 1
	}

	return replicas - random
}

// getDeploymentReplicas gets the corresponding Deployment replica count through Pod's OwnerReference
func getDeploymentReplicas(pod *corev1.Pod) int32 {
	// Iterate through Pod's OwnerReferences
	for _, owner := range pod.OwnerReferences {
		klog.Infof("Pod owner: %s, kind: %s", owner.Name, owner.Kind)
		if owner.Kind == "ReplicaSet" {
			// Get ReplicaSet
			rs, err := kube.KubeClientset().AppsV1().ReplicaSets(pod.Namespace).Get(
				context.TODO(),
				owner.Name,
				metav1.GetOptions{},
			)
			if err != nil {
				klog.Errorf("Failed to get ReplicaSet %s: %v", owner.Name, err)
				continue
			}

			// Iterate through ReplicaSet's OwnerReferences to get Deployment
			for _, rsOwner := range rs.OwnerReferences {
				klog.Infof("ReplicaSet owner: %s, kind: %s", rsOwner.Name, rsOwner.Kind)
				if rsOwner.Kind == "Deployment" {
					deployment, err := kube.KubeClientset().AppsV1().Deployments(pod.Namespace).Get(
						context.TODO(),
						rsOwner.Name,
						metav1.GetOptions{},
					)
					if err != nil {
						klog.Errorf("Failed to get Deployment %s: %v", rsOwner.Name, err)
						continue
					}

					// Return Deployment replica count
					if deployment.Spec.Replicas != nil {
						klog.Infof("Found Deployment %s with %d replicas for Pod %s",
							rsOwner.Name, *deployment.Spec.Replicas, pod.Name)
						return *deployment.Spec.Replicas
					}
					return 1 // Default replica count
				}
			}
		}
	}

	// If no Deployment found, return default value
	klog.Infof("No Deployment found for Pod %s, using default replica count", pod.Name)
	return 1
}

func getSafeLabels(podLabels map[string]string) map[string]string {
	safeLabels := make(map[string]string)
	businessLabels := []string{"app", "version", "component", "tier", "env"}

	if podLabels != nil {
		for _, key := range businessLabels {
			if value, exists := podLabels[key]; exists {
				safeLabels[key] = value
			}
		}
	}

	if len(safeLabels) == 0 {
		if appValue, exists := podLabels["app"]; exists {
			safeLabels["app"] = appValue
		} else {
			safeLabels["topology-group"] = "default"
		}
	}

	return safeLabels
}

func ToV1AdmissionResponse(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
