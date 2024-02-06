package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
type DataRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataRestoreSpec `json:"spec"`
	Status RestoreStatus   `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DataRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DataRestore `json:"items"`
}

type DataRestoreSpec struct {
	VolumeSnapshotClass string `json:"volumeSnapshotClass"`
	Storage             string `json:"storage,omitempty"`
	SnapshotCRName      string `json:"snapshotCRName"`
}

type RestoreStatus struct {
	Progress string `json:"progress,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
type DataSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataSnapshotSpec `json:"spec"`
	Status SnapshotStatus   `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DataSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DataSnapshot `json:"items"`
}

type DataSnapshotSpec struct {
	VolumeSnapshotClass string `json:"volumeSnapshotClass"`
	PvcName             string `json:"pvcName"`
	Namespace           string `json:"namespace"`
}

type SnapshotStatus struct {
	Progress     string `json:"progress,omitempty"`
	SnapshotName string `json:"snapshotName,omitempty"`
}
