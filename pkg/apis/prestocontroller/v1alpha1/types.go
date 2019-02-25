package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Presto is a specification for a Presto resource
type Presto struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrestoSpec   `json:"spec"`
	Status PrestoStatus `json:"status"`
}

// PrestoSpec is the spec for a Presto resource
type PrestoSpec struct {
	ClusterName       string `json:"clusterName"`
	Replicas          *int32 `json:"replicas"`
	CoordinatorConfig string `json:"coordinatorConfig"`
	WorkerConfig      string `json:"workerConfig"`
	CatalogConfig     string `json:"catalogConfig"`
}

// PrestoStatus is the status for a Presto resource
type PrestoStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrestoList is a list of Presto resources
type PrestoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Presto `json:"items"`
}
