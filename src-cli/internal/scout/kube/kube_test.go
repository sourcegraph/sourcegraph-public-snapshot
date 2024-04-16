package kube

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAcceptedFileSystem(t *testing.T) {
	cases := []struct {
		name       string
		filesystem string
		want       bool
	}{
		{
			name:       "should return true if filesystem matches 'matched' regular expression",
			filesystem: "/dev/sda",
			want:       true,
		},
		{
			name:       "should return false if filesystem doesn't match 'matched' regular expression",
			filesystem: "/dev/sda1",
			want:       false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := acceptedFileSystem(tc.filesystem)
			if got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestGetPod(t *testing.T) {
	cases := []struct {
		name    string
		podList []corev1.Pod
		wantPod string
	}{
		{
			name: "should return correct pod",
			podList: []corev1.Pod{
				*testPod("sg", "soucegraph-frontend-0", "sourcegraph-frontend"),
				*testPod("sg", "gitserver-0", "gitserver"),
				*testPod("sg", "indexed-search-0", "indexed-search"),
			},
			wantPod: "gitserver-0",
		},
		{
			name: "should return empty pod if pod not found",
			podList: []corev1.Pod{
				*testPod("sg", "soucegraph-frontend-0", "sourcegraph-frontend"),
				*testPod("sg", "indexed-search-0", "indexed-search"),
			},
			wantPod: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, _ := GetPod("gitserver-0", tc.podList)
			gotPod := got.Name

			if gotPod != tc.wantPod {
				t.Errorf("want pod %s, got pod %s", tc.wantPod, gotPod)
			}
		})
	}
}

func testPod(namespace, podName, containerName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      podName,
		},
	}
}
