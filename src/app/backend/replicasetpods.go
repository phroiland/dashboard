// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"sort"
)

// TotalRestartCountSorter sorts ReplicaSetPodWithContainers by restarts number.
type TotalRestartCountSorter []ReplicaSetPodWithContainers

func (a TotalRestartCountSorter) Len() int      { return len(a) }
func (a TotalRestartCountSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a TotalRestartCountSorter) Less(i, j int) bool {
	return a[i].TotalRestartCount > a[j].TotalRestartCount
}

// Information about a Container that belongs to a Pod.
type PodContainer struct {
	// Name of a Container.
	Name string `json:"name"`

	// Number of restarts.
	RestartCount int `json:"restartCount"`
}

// List of pods that belongs to a Replica Set.
type ReplicaSetPods struct {
	// List of pods that belongs to a Replica Set.
	Pods []ReplicaSetPodWithContainers `json:"pods"`
}

// Detailed information about a Pod that belongs to a Replica Set.
type ReplicaSetPodWithContainers struct {
	// Name of the Pod.
	Name string `json:"name"`

	// Time the Pod has started. Empty if not started.
	StartTime *unversioned.Time `json:"startTime"`

	// Total number of restarts.
	TotalRestartCount int `json:"totalRestartCount"`

	// List of Containers that belongs to particular Pod.
	PodContainers []PodContainer `json:"podContainers"`
}

// Returns list of pods with containers for the given replica set in the given namespace.
// Limit specify the number of records to return. There is no limit when given value is zero.
func GetReplicaSetPods(client *client.Client, namespace string, name string, limit int) (
	*ReplicaSetPods, error) {
	pods, err := getRawReplicaSetPods(client, namespace, name)
	if err != nil {
		return nil, err
	}

	return getReplicaSetPods(pods.Items, limit), nil
}

// Creates and return structure containing pods with containers for given replica set.
// Data is sorted by total number of restarts for replica set pod.
// Result set can be limited
func getReplicaSetPods(pods []api.Pod, limit int) *ReplicaSetPods {
	replicaSetPods := &ReplicaSetPods{}
	for _, pod := range pods {
		totalRestartCount := 0
		replicaSetPodWithContainers := ReplicaSetPodWithContainers{
			Name:      pod.Name,
			StartTime: pod.Status.StartTime,
		}
		for _, containerStatus := range pod.Status.ContainerStatuses {
			podContainer := PodContainer{
				Name:         containerStatus.Name,
				RestartCount: containerStatus.RestartCount,
			}
			replicaSetPodWithContainers.PodContainers =
				append(replicaSetPodWithContainers.PodContainers, podContainer)
			totalRestartCount += containerStatus.RestartCount
		}
		replicaSetPodWithContainers.TotalRestartCount = totalRestartCount
		replicaSetPods.Pods = append(replicaSetPods.Pods, replicaSetPodWithContainers)
	}
	sort.Sort(TotalRestartCountSorter(replicaSetPods.Pods))

	if limit > 0 {
		if limit > len(replicaSetPods.Pods) {
			limit = len(replicaSetPods.Pods)
		}
		replicaSetPods.Pods = replicaSetPods.Pods[0:limit]
	}
	return replicaSetPods
}