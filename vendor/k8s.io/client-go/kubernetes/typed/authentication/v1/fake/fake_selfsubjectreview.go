/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testing "k8s.io/client-go/testing"
)

// FakeSelfSubjectReviews implements SelfSubjectReviewInterface
type FakeSelfSubjectReviews struct {
	Fake *FakeAuthenticationV1
}

var selfsubjectreviewsResource = v1.SchemeGroupVersion.WithResource("selfsubjectreviews")

var selfsubjectreviewsKind = v1.SchemeGroupVersion.WithKind("SelfSubjectReview")

// Create takes the representation of a selfSubjectReview and creates it.  Returns the server's representation of the selfSubjectReview, and an error, if there is any.
func (c *FakeSelfSubjectReviews) Create(ctx context.Context, selfSubjectReview *v1.SelfSubjectReview, opts metav1.CreateOptions) (result *v1.SelfSubjectReview, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(selfsubjectreviewsResource, selfSubjectReview), &v1.SelfSubjectReview{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1.SelfSubjectReview), err
}
