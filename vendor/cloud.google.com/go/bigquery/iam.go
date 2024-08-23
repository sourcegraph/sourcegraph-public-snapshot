// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bigquery

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/iam"
	"cloud.google.com/go/iam/apiv1/iampb"
	"cloud.google.com/go/internal/trace"
	bq "google.golang.org/api/bigquery/v2"
)

// IAM provides access to an iam.Handle that allows access to IAM functionality for
// the given BigQuery table.  For more information, see
// https://pkg.go.dev/cloud.google.com/go/iam
func (t *Table) IAM() *iam.Handle {
	return iam.InternalNewHandleClient(&bqIAMClient{
		bqs: t.c.bqs,
	}, fmt.Sprintf("projects/%s/datasets/%s/tables/%s",
		t.ProjectID, t.DatasetID, t.TableID))
}

// bqIAMClient is a client that satisfies the IAM "client" interface.
//
// This client has an explicit assumption that it's only working with Table resources.
type bqIAMClient struct {
	bqs *bq.Service
}

func (c *bqIAMClient) Get(ctx context.Context, resource string) (p *iampb.Policy, err error) {
	return c.GetWithVersion(ctx, resource, 1)
}

func (c *bqIAMClient) GetWithVersion(ctx context.Context, resource string, requestedPolicyVersion int32) (p *iampb.Policy, err error) {
	if requestedPolicyVersion > 1 {
		return nil, errors.New("bigquery: only IAM policy version 1 is supported")
	}
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.IAM.Get")
	defer func() { trace.EndSpan(ctx, err) }()

	iamReq := &bq.GetIamPolicyRequest{
		Options: &bq.GetPolicyOptions{
			RequestedPolicyVersion: int64(requestedPolicyVersion),
		},
	}
	call := c.bqs.Tables.GetIamPolicy(resource, iamReq).Context(ctx)
	setClientHeader(call.Header())

	var bqp *bq.Policy
	err = runWithRetry(ctx, func() error {
		bqp, err = call.Do()
		return err
	})
	if err != nil {
		return nil, err
	}
	return iamFromBigQueryPolicy(bqp), nil
}

func (c *bqIAMClient) Set(ctx context.Context, resource string, p *iampb.Policy) (err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.IAM.Set")
	defer func() { trace.EndSpan(ctx, err) }()

	bqp := iamToBigQueryPolicy(p)
	call := c.bqs.Tables.SetIamPolicy(resource, &bq.SetIamPolicyRequest{Policy: bqp}).Context(ctx)
	setClientHeader(call.Header())
	return runWithRetry(ctx, func() error {
		_, err := call.Do()
		return err
	})
}

func (c *bqIAMClient) Test(ctx context.Context, resource string, perms []string) (p []string, err error) {
	ctx = trace.StartSpan(ctx, "cloud.google.com/go/bigquery.IAM.Test")
	defer func() { trace.EndSpan(ctx, err) }()

	call := c.bqs.Tables.TestIamPermissions(resource, &bq.TestIamPermissionsRequest{Permissions: perms}).Context(ctx)
	setClientHeader(call.Header())

	var res *bq.TestIamPermissionsResponse
	err = runWithRetry(ctx, func() error {
		res, err = call.Do()
		return err
	})
	if err != nil {
		return nil, err
	}
	return res.Permissions, nil
}

func iamToBigQueryPolicy(ip *iampb.Policy) *bq.Policy {
	return &bq.Policy{
		Bindings: iamToBigQueryBindings(ip.Bindings),
		Etag:     string(ip.Etag),
		Version:  int64(ip.Version),
	}
}

func iamToBigQueryBindings(ibs []*iampb.Binding) []*bq.Binding {
	var bqBindings []*bq.Binding
	for _, ib := range ibs {
		bqBindings = append(bqBindings, &bq.Binding{
			Role:    ib.Role,
			Members: ib.Members,
		})
	}
	return bqBindings
}

func iamFromBigQueryPolicy(bqPolicy *bq.Policy) *iampb.Policy {
	return &iampb.Policy{
		Bindings: iamFromBigQueryBindings(bqPolicy.Bindings),
		Etag:     []byte(bqPolicy.Etag),
		Version:  int32(bqPolicy.Version),
	}
}

func iamFromBigQueryBindings(bqBindings []*bq.Binding) []*iampb.Binding {
	var ibs []*iampb.Binding
	for _, bqb := range bqBindings {
		ibs = append(ibs, &iampb.Binding{
			Role:    bqb.Role,
			Members: bqb.Members,
		})
	}
	return ibs
}
