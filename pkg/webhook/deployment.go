package webhook

import (
	"fmt"

	v1 "k8s.io/api/admission/v1"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog/v2"
)

var DeploymentReplicaSetNum = make(chan int32, 1000)

func admitDeployments(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.V(2).Info("admitting deployments")
	deploymentResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "deployments"}

	if ar.Request.Resource != deploymentResource {
		err := fmt.Errorf("expect resource to be %s", deploymentResource)
		klog.Error(err)
		return ToV1AdmissionResponse(err)
	}

	raw := ar.Request.Object.Raw
	deployment := apps.Deployment{}
	deserializer := serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &deployment); err != nil {
		klog.Error(err)
		return ToV1AdmissionResponse(err)
	}

	replicas := deployment.Spec.Replicas
	if replicas == nil {
		DeploymentReplicaSetNum <- 1
	} else {
		DeploymentReplicaSetNum <- *replicas
	}

	reviewResponse := v1.AdmissionResponse{}
	reviewResponse.Allowed = true
	return &reviewResponse
}
