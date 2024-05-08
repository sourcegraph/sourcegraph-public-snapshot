// gqltest prov
// copy from: https://github.com/sourcegraph/controller/tree/main/internal/srcgql/gqltest

package gqltest

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type MakeRequestStub func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error

// MakeRequestResultStub generates a stub that implements MakeResult in a
// gqltest.FakeClient to always return mockRespData as the GraphQL response data.
// mockRespData should be a value instance of the generated response type, for example:
//
//	gqlClient := &gqltest.FakeClient{
//		MakeRequestStub: gqltest.MakeRequestResultStub(srcgql.GetSiteConfigurationResponse{
//			Site: srcgql.GetSiteConfigurationSite{
//				Configuration: srcgql.GetSiteConfigurationSiteConfiguration{
//					EffectiveContents: `{
//						"email.address": "robert@sourcegraph.com",
//						"email.smtp": {
//							"host": "bobheadxi.dev"
//						}
//					}`,
//				},
//			},
//		}),
//	}
//
// Multiple result stubs can be composed with MakeRequestStubInvocations to mock multiple
// GraphQL requests:
//
//	gqlClient := &gqltest.FakeClient{
//		MakeRequestStub: gqltest.MakeRequestStubInvocations(
//			gqltest.MakeRequestResultStub(...),
//			gqltest.MakeRequestResultStub(...),
//			func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
//				return errors.New("oh no!")
//			}
//		),
//	}
//
// See MakeRequestStubInvocations for more details.
func MakeRequestResultStub[T any](mockRespData T) MakeRequestStub {
	return func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
		switch d := resp.Data.(type) {
		case *T:
			*d = mockRespData
		default:
			return errors.Newf(`got unexpected operation %q with data type "%T", faker wanted "%T"`, req.OpName, resp.Data, mockRespData)
		}
		return nil
	}
}

// MakeRequestErrorStub generates a stub that returns an error
func MakeRequestResultErrStub(err error) MakeRequestStub {
	return func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
		return err
	}
}

// MakeRequestStubInvocations can be used to use different stubs for each invocation.
// Stubs are used in order of invocation.
//
// Invocation is 1-indexed
func MakeRequestStubInvocations(stubs ...MakeRequestStub) MakeRequestStub {
	var invocation int
	return func(ctx context.Context, req *graphql.Request, resp *graphql.Response) error {
		if invocation > len(stubs) {
			return errors.Newf("unexpected operation %q with data type %T, invocation %d", req.OpName, resp.Data, invocation)
		}
		invocation += 1
		err := stubs[invocation-1](ctx, req, resp)
		if err != nil {
			return errors.Wrapf(err, "invocation %d", invocation)
		}
		return nil
	}
}

// UnmarshalRequestVariables unmarshals the variables of a GraphQL variables into a map
// that can be used for golden testing.
//
// genql input variables are not exported structs, so we cannot use json.Unmarshal directly.
func UnmarshalVariables(t *testing.T, req *graphql.Request) map[string]any {
	require.NotNil(t, req)
	require.NotNil(t, req.Variables)

	bytes, err := json.Marshal(req.Variables)
	require.NoErrorf(t, err, "marshal variable: %v", req.Variables)

	var m map[string]any
	err = json.Unmarshal(bytes, &m)
	require.NoErrorf(t, err, "unmarshal variable: %v", req.Variables)

	return m
}

type Request struct {
	OpName    string         `json:"operationName"`
	Variables map[string]any `json:"variables"`
}

// UnmarshalRequest unmarshals a GraphQL request into a map that can be used for golden testing.
func UnmarshalRequest(t *testing.T, req *graphql.Request) Request {
	require.NotNil(t, req)
	result := Request{
		OpName: req.OpName,
	}
	if req.Variables != nil {
		result.Variables = UnmarshalVariables(t, req)
	}
	return result
}

func (fake *FakeClient) MakeRequestArgsGraphqlRequestForCall(i int) *graphql.Request {
	fake.makeRequestMutex.RLock()
	defer fake.makeRequestMutex.RUnlock()
	argsForCall := fake.makeRequestArgsForCall[i]
	return argsForCall.arg2
}
