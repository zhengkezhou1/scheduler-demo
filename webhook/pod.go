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

	// 创建 JSON Patch 来添加 Affinity 和 TopologySpreadConstraints
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
										Values:   []string{"on-demand"},
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
				{
					MaxSkew:           1,
					TopologyKey:       "node.kubernetes.io/capacity",
					WhenUnsatisfiable: corev1.ScheduleAnyway,
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
	// 只选择关键的业务标签，避免使用系统生成的标签
	safeLabels := make(map[string]string)

	// 常用的业务标签
	businessLabels := []string{"app", "version", "component", "tier", "env"}

	if podLabels != nil {
		for _, key := range businessLabels {
			if value, exists := podLabels[key]; exists {
				safeLabels[key] = value
			}
		}
	}

	// 如果没有找到任何业务标签，使用 app 标签或创建一个默认标签
	if len(safeLabels) == 0 {
		if appValue, exists := podLabels["app"]; exists {
			safeLabels["app"] = appValue
		} else {
			// 创建一个基于 pod 名称的标签作为后备
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
