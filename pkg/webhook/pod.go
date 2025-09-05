package webhook

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"
)

func admitPods(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.V(2).Info("admitting pods")
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

	if ar.Request.Resource != podResource {
		err := fmt.Errorf("expect resource to be %s", podResource)
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}

	raw := ar.Request.Object.Raw
	pod := corev1.Pod{}
	deserializer := serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
		klog.Error(err)
		return toV1AdmissionResponse(err)
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
		{
			"op":   "add",
			"path": "/spec/topologySpreadConstraints",
			"value": []corev1.TopologySpreadConstraint{
				// TODO: MaxSkew value depends on how many nodes we have.
				{
					MaxSkew:           30,
					TopologyKey:       "node.kubernetes.io/capacity",
					WhenUnsatisfiable: corev1.DoNotSchedule,
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: getSafeLabels(pod.Labels),
					},
				},
			},
		},
	}

	patchBytes, err := json.Marshal(patches)
	if err != nil {
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}

	pt := v1.PatchTypeJSONPatch
	reviewResponse.PatchType = &pt
	reviewResponse.Patch = patchBytes
	return &reviewResponse
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

func toV1AdmissionResponse(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
