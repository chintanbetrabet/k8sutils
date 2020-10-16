package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CustomAutoscalerSpec defines the desired state of CustomAutoscaler
type CustomAutoscalerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	DeploymentName  string
	MinReplicas     int32
	MaxReplicas     int32
	DefaultReplicas int32
	ScalingPolicy   string
	Metrics         []CustomAutoscalerMetric `json:"items"`
}

// CustomAutoscalerStatus defines the observed state of CustomAutoscaler
type CustomAutoscalerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Replicas      int32
	LastUpscale   CustomDateTime
	LastDownscale CustomDateTime
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CustomAutoscaler is the Schema for the customautoscalers API
// +k8s:openapi-gen=true
type CustomAutoscaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomAutoscalerSpec   `json:"spec,omitempty"`
	Status CustomAutoscalerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CustomAutoscalerList contains a list of CustomAutoscaler
type CustomAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomAutoscaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomAutoscaler{}, &CustomAutoscalerList{})
}

type CustomAutoscalerMetric struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Name                     string
	targetAverageUtilization int
}

type CustomDateTime struct {
	Time time.Time
}

func (in *CustomDateTime) DeepCopyInto(out *CustomDateTime) {
	*out = *in
	out.Time = in.Time
}

func (autoscaler *CustomAutoscalerSpec) PopulateDefaults() {
	if autoscaler.ScalingPolicy == "" {
		autoscaler.ScalingPolicy = "default"
	}
	if autoscaler.DefaultReplicas == 0 {
		autoscaler.DefaultReplicas = 1
	}
}
