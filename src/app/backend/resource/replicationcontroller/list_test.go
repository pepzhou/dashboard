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

package replicationcontroller

import (
	"reflect"
	"testing"

	"github.com/kubernetes/dashboard/src/app/backend/api"
	"github.com/kubernetes/dashboard/src/app/backend/resource/common"
	"github.com/kubernetes/dashboard/src/app/backend/resource/dataselect"
	"github.com/kubernetes/dashboard/src/app/backend/resource/metric"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

func TestGetReplicationControllerList(t *testing.T) {
	replicas := int32(0)
	events := []v1.Event{}
	controller := true
	firstAppOwnerRef := []metaV1.OwnerReference{{
		Kind:       "ReplicationController",
		Name:       "my-name-1",
		UID:        "uid-1",
		Controller: &controller,
	}}

	cases := []struct {
		replicationControllers []v1.ReplicationController
		services               []v1.Service
		pods                   []v1.Pod
		nodes                  []v1.Node
		expected               *ReplicationControllerList
	}{
		{nil, nil, nil, nil,
			&ReplicationControllerList{
				ReplicationControllers: []ReplicationController{},
				CumulativeMetrics:      make([]metric.Metric, 0),
			},
		},
		{
			[]v1.ReplicationController{
				{
					ObjectMeta: metaV1.ObjectMeta{
						Name:      "my-app-1",
						Namespace: "namespace-1",
						UID:       "uid-1",
					},
					Spec: v1.ReplicationControllerSpec{
						Replicas: &replicas,
						Selector: map[string]string{"app": "my-name-1"},
						Template: &v1.PodTemplateSpec{
							Spec: v1.PodSpec{Containers: []v1.Container{{Image: "my-container-image-1"}}},
						},
					},
				},
				{
					ObjectMeta: metaV1.ObjectMeta{
						Name:      "my-app-2",
						Namespace: "namespace-2",
						UID:       "uid-2",
					},
					Spec: v1.ReplicationControllerSpec{
						Replicas: &replicas,
						Selector: map[string]string{"app": "my-name-2", "ver": "2"},
						Template: &v1.PodTemplateSpec{
							Spec: v1.PodSpec{Containers: []v1.Container{{Image: "my-container-image-2"}}},
						},
					},
				},
			},
			[]v1.Service{
				{
					Spec: v1.ServiceSpec{Selector: map[string]string{"app": "my-name-1"}},
					ObjectMeta: metaV1.ObjectMeta{
						Name:      "my-app-1",
						Namespace: "namespace-1",
						UID:       "uid-1",
					},
				},
				{
					Spec: v1.ServiceSpec{Selector: map[string]string{"app": "my-name-2", "ver": "2"}},
					ObjectMeta: metaV1.ObjectMeta{
						Name:      "my-app-2",
						Namespace: "namespace-2",
						UID:       "uid-1",
					},
				},
			},
			[]v1.Pod{
				{
					ObjectMeta: metaV1.ObjectMeta{
						Namespace:       "namespace-1",
						OwnerReferences: firstAppOwnerRef,
					},
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				},
				{
					ObjectMeta: metaV1.ObjectMeta{
						Namespace:       "namespace-1",
						OwnerReferences: firstAppOwnerRef,
					},
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				},
				{
					ObjectMeta: metaV1.ObjectMeta{
						Namespace:       "namespace-1",
						OwnerReferences: firstAppOwnerRef,
					},
					Status: v1.PodStatus{
						Phase: v1.PodPending,
					},
				},
				{
					ObjectMeta: metaV1.ObjectMeta{
						Namespace:       "namespace-2",
						OwnerReferences: firstAppOwnerRef,
					},
					Status: v1.PodStatus{
						Phase: v1.PodPending,
					},
				},
				{
					ObjectMeta: metaV1.ObjectMeta{
						Namespace:       "namespace-1",
						OwnerReferences: firstAppOwnerRef,
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
				{
					ObjectMeta: metaV1.ObjectMeta{
						Namespace:       "namespace-1",
						OwnerReferences: firstAppOwnerRef,
					},
					Status: v1.PodStatus{
						Phase: v1.PodSucceeded,
					},
				},
				{
					ObjectMeta: metaV1.ObjectMeta{
						Namespace:       "namespace-1",
						OwnerReferences: firstAppOwnerRef,
					},
					Status: v1.PodStatus{
						Phase: v1.PodUnknown,
					},
				},
			},
			[]v1.Node{{
				Status: v1.NodeStatus{
					Addresses: []v1.NodeAddress{
						{
							Type:    v1.NodeExternalIP,
							Address: "192.168.1.108",
						},
					},
				},
			},
			},
			&ReplicationControllerList{
				ListMeta:          api.ListMeta{TotalItems: 2},
				CumulativeMetrics: make([]metric.Metric, 0),
				ReplicationControllers: []ReplicationController{
					{
						ObjectMeta: api.ObjectMeta{
							Name:      "my-app-1",
							Namespace: "namespace-1",
						},
						TypeMeta:        api.TypeMeta{Kind: api.ResourceKindReplicationController},
						ContainerImages: []string{"my-container-image-1"},
						Pods: common.PodInfo{
							Failed:    2,
							Pending:   1,
							Running:   1,
							Succeeded: 1,
							Warnings:  []common.Event{},
						},
					}, {
						ObjectMeta: api.ObjectMeta{
							Name:      "my-app-2",
							Namespace: "namespace-2",
						},
						TypeMeta:        api.TypeMeta{Kind: api.ResourceKindReplicationController},
						ContainerImages: []string{"my-container-image-2"},
						Pods: common.PodInfo{
							Warnings: []common.Event{},
						},
					},
				},
			},
		},
	}
	for _, c := range cases {
		actual := CreateReplicationControllerList(c.replicationControllers, dataselect.NoDataSelect,
			c.pods, events, nil)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("getReplicationControllerList(%#v, %#v) == \n%#v\nexpected \n%#v\n",
				c.replicationControllers, c.services, actual, c.expected)
		}
	}
}
