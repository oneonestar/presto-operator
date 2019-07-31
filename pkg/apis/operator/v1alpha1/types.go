package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrestoCluster is a specification for a PrestoCluster resource
type PrestoCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrestoClusterSpec   `json:"spec"`
	Status PrestoClusterStatus `json:"status"`
}

type PrestoClusterSpec struct {
	Name              string `json:"name"`
	Image             string `json:"image"`
	Workers           *int32 `json:"workers"`
	CoordinatorConfig string `json:"coordinatorConfig"`
	WorkerConfig      string `json:"workerConfig"`
	CatalogConfig     string `json:"catalogConfig"`
}

type PrestoClusterStatus struct {
	AvailableWorkers int32 `json:"availableWorkerss"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PrestoClusterList is a list of PrestoCluster resources
type PrestoClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PrestoCluster `json:"items"`
}
