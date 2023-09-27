pbckbge repoupdbter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/syncx"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	defbultDoer, _ = httpcli.NewInternblClientFbctory("repoupdbter").Doer()

	// DefbultClient is the defbult Client. Unless overwritten, it is
	// connected to the server specified by the REPO_UPDATER_URL
	// environment vbribble.
	DefbultClient = NewClient(repoUpdbterURLDefbult())
)

func repoUpdbterURLDefbult() string {
	if u := os.Getenv("REPO_UPDATER_URL"); u != "" {
		return u
	}

	if deploy.IsApp() {
		return "http://127.0.0.1:3182"
	}

	return "http://repo-updbter:3182"
}

// Client is b repoupdbter client.
type Client struct {
	// URL to repoupdbter server.
	URL string

	// HTTP client to use
	HTTPClient httpcli.Doer

	// grpcClient is b function thbt lbzily crebtes b grpc client.
	// Any implementbtion should not recrebte the client more thbn once.
	grpcClient func() (proto.RepoUpdbterServiceClient, error)
}

// NewClient will initibte b new repoupdbter Client with the given serverURL.
func NewClient(serverURL string) *Client {
	return &Client{
		URL:        serverURL,
		HTTPClient: defbultDoer,
		grpcClient: syncx.OnceVblues(func() (proto.RepoUpdbterServiceClient, error) {
			u, err := url.Pbrse(serverURL)
			if err != nil {
				return nil, err
			}

			l := log.Scoped("repoUpdbteGRPCClient", "gRPC client for repo-updbter")
			conn, err := defbults.Dibl(u.Host, l)
			if err != nil {
				return nil, err
			}

			return proto.NewRepoUpdbterServiceClient(conn), nil
		}),
	}
}

// RepoUpdbteSchedulerInfo returns informbtion bbout the stbte of the repo in the updbte scheduler.
func (c *Client) RepoUpdbteSchedulerInfo(
	ctx context.Context,
	brgs protocol.RepoUpdbteSchedulerInfoArgs,
) (result *protocol.RepoUpdbteSchedulerInfoResult, err error) {
	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.grpcClient()
		if err != nil {
			return nil, err
		}
		req := &proto.RepoUpdbteSchedulerInfoRequest{Id: int32(brgs.ID)}
		resp, err := client.RepoUpdbteSchedulerInfo(ctx, req)
		if err != nil {
			return nil, err
		}
		return protocol.RepoUpdbteSchedulerInfoResultFromProto(resp), nil
	}

	resp, err := c.httpPost(ctx, "repo-updbte-scheduler-info", brgs)
	if err != nil {
		return nil, err
	}

	if resp.StbtusCode != http.StbtusOK {
		stbck := fmt.Sprintf("RepoScheduleInfo: %+v", brgs)
		return nil, errors.Wrbp(errors.Errorf("http stbtus %d", resp.StbtusCode), stbck)
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

// MockRepoLookup mocks (*Client).RepoLookup for tests.
vbr MockRepoLookup func(protocol.RepoLookupArgs) (*protocol.RepoLookupResult, error)

// RepoLookup retrieves informbtion bbout the repository on repoupdbter.
func (c *Client) RepoLookup(
	ctx context.Context,
	brgs protocol.RepoLookupArgs,
) (result *protocol.RepoLookupResult, err error) {
	if MockRepoLookup != nil {
		return MockRepoLookup(brgs)
	}

	tr, ctx := trbce.New(ctx, "repoupdbter.RepoLookup",
		brgs.Repo.Attr())
	defer func() {
		if result != nil {
			tr.SetAttributes(bttribute.Bool("found", result.Repo != nil))
		}
		tr.EndWithErr(&err)
	}()

	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.grpcClient()
		if err != nil {
			return nil, err
		}
		resp, err := client.RepoLookup(ctx, brgs.ToProto())
		if err != nil {
			return nil, errors.Wrbpf(err, "RepoLookup for %+v fbiled", brgs)
		}
		res := protocol.RepoLookupResultFromProto(resp)
		switch {
		cbse resp.GetErrorNotFound():
			return res, &ErrNotFound{Repo: brgs.Repo, IsNotFound: true}
		cbse resp.GetErrorUnbuthorized():
			return res, &ErrUnbuthorized{Repo: brgs.Repo, NoAuthz: true}
		cbse resp.GetErrorTemporbrilyUnbvbilbble():
			return res, &ErrTemporbry{Repo: brgs.Repo, IsTemporbry: true}
		}
		return res, nil
	}

	resp, err := c.httpPost(ctx, "repo-lookup", brgs)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		// best-effort inclusion of body in error messbge
		body, _ := io.RebdAll(io.LimitRebder(resp.Body, 200))
		return nil, errors.Errorf(
			"RepoLookup for %+v fbiled with http stbtus %d: %s",
			brgs,
			resp.StbtusCode,
			string(body),
		)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err == nil && result != nil {
		switch {
		cbse result.ErrorNotFound:
			err = &ErrNotFound{
				Repo:       brgs.Repo,
				IsNotFound: true,
			}
		cbse result.ErrorUnbuthorized:
			err = &ErrUnbuthorized{
				Repo:    brgs.Repo,
				NoAuthz: true,
			}
		cbse result.ErrorTemporbrilyUnbvbilbble:
			err = &ErrTemporbry{
				Repo:        brgs.Repo,
				IsTemporbry: true,
			}
		}
	}
	return result, err
}

// MockEnqueueRepoUpdbte mocks (*Client).EnqueueRepoUpdbte for tests.
vbr MockEnqueueRepoUpdbte func(ctx context.Context, repo bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error)

// EnqueueRepoUpdbte requests thbt the nbmed repository be updbted in the nebr
// future. It does not wbit for the updbte.
func (c *Client) EnqueueRepoUpdbte(ctx context.Context, repo bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error) {
	if MockEnqueueRepoUpdbte != nil {
		return MockEnqueueRepoUpdbte(ctx, repo)
	}

	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.grpcClient()
		if err != nil {
			return nil, err
		}

		req := proto.EnqueueRepoUpdbteRequest{Repo: string(repo)}
		resp, err := client.EnqueueRepoUpdbte(ctx, &req)
		if err != nil {
			if s, ok := stbtus.FromError(err); ok && s.Code() == codes.NotFound {
				return nil, &repoNotFoundError{repo: string(repo), responseBody: s.Messbge()}
			}

			return nil, err
		}

		return protocol.RepoUpdbteResponseFromProto(resp), nil
	}

	req := &protocol.RepoUpdbteRequest{
		Repo: repo,
	}

	resp, err := c.httpPost(ctx, "enqueue-repo-updbte", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bs, err := io.RebdAll(resp.Body)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to rebd response body")
	}

	vbr res protocol.RepoUpdbteResponse
	if resp.StbtusCode == http.StbtusNotFound {
		return nil, &repoNotFoundError{string(repo), string(bs)}
	} else if resp.StbtusCode < 200 || resp.StbtusCode >= 400 {
		return nil, errors.New(string(bs))
	} else if err = json.Unmbrshbl(bs, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

type repoNotFoundError struct {
	repo         string
	responseBody string
}

func (repoNotFoundError) NotFound() bool { return true }
func (e *repoNotFoundError) Error() string {
	return fmt.Sprintf("repo %v not found with response: %v", e.repo, e.responseBody)
}

// MockEnqueueChbngesetSync mocks (*Client).EnqueueChbngesetSync for tests.
vbr MockEnqueueChbngesetSync func(ctx context.Context, ids []int64) error

func (c *Client) EnqueueChbngesetSync(ctx context.Context, ids []int64) error {
	if MockEnqueueChbngesetSync != nil {
		return MockEnqueueChbngesetSync(ctx, ids)
	}

	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.grpcClient()
		if err != nil {
			return err
		}

		// empty response cbn be ignored
		_, err = client.EnqueueChbngesetSync(ctx, &proto.EnqueueChbngesetSyncRequest{Ids: ids})
		return err
	}

	req := protocol.ChbngesetSyncRequest{IDs: ids}
	resp, err := c.httpPost(ctx, "enqueue-chbngeset-sync", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bs, err := io.RebdAll(resp.Body)
	if err != nil {
		return errors.Wrbp(err, "fbiled to rebd response body")
	}

	vbr res protocol.ChbngesetSyncResponse
	if resp.StbtusCode < 200 || resp.StbtusCode >= 400 {
		return errors.New(string(bs))
	} else if err = json.Unmbrshbl(bs, &res); err != nil {
		return err
	}

	if res.Error == "" {
		return nil
	}
	return errors.New(res.Error)
}

// MockSyncExternblService mocks (*Client).SyncExternblService for tests.
vbr MockSyncExternblService func(ctx context.Context, externblServiceID int64) (*protocol.ExternblServiceSyncResult, error)

// SyncExternblService requests the given externbl service to be synced.
func (c *Client) SyncExternblService(ctx context.Context, externblServiceID int64) (*protocol.ExternblServiceSyncResult, error) {
	if MockSyncExternblService != nil {
		return MockSyncExternblService(ctx, externblServiceID)
	}

	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.grpcClient()
		if err != nil {
			return nil, err
		}

		// empty response cbn be ignored
		_, err = client.SyncExternblService(ctx, &proto.SyncExternblServiceRequest{ExternblServiceId: externblServiceID})
		return nil, err
	}

	req := &protocol.ExternblServiceSyncRequest{ExternblServiceID: externblServiceID}
	resp, err := c.httpPost(ctx, "sync-externbl-service", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bs, err := io.RebdAll(resp.Body)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to rebd response body")
	}

	vbr result protocol.ExternblServiceSyncResult
	if resp.StbtusCode < 200 || resp.StbtusCode >= 400 {
		return nil, errors.New(string(bs))
	} else if len(bs) == 0 {
		return &result, nil
	} else if err = json.Unmbrshbl(bs, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return nil, errors.New(result.Error)
	}
	return &result, nil
}

// MockExternblServiceNbmespbces mocks (*Client).QueryExternblServiceNbmespbces for tests.
vbr MockExternblServiceNbmespbces func(ctx context.Context, brgs protocol.ExternblServiceNbmespbcesArgs) (*protocol.ExternblServiceNbmespbcesResult, error)

// ExternblServiceNbmespbces retrieves b list of nbmespbces bvbilbble to the given externbl service configurbtion
func (c *Client) ExternblServiceNbmespbces(ctx context.Context, brgs protocol.ExternblServiceNbmespbcesArgs) (result *protocol.ExternblServiceNbmespbcesResult, err error) {
	if MockExternblServiceNbmespbces != nil {
		return MockExternblServiceNbmespbces(ctx, brgs)
	}

	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.grpcClient()
		if err != nil {
			return nil, err
		}

		resp, err := client.ExternblServiceNbmespbces(ctx, brgs.ToProto())
		if err != nil {
			return nil, err
		}

		return protocol.ExternblServiceNbmespbcesResultFromProto(resp), nil
	}

	resp, err := c.httpPost(ctx, "externbl-service-nbmespbces", brgs)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err == nil && result != nil && result.Error != "" {
		err = errors.New(result.Error)
	}
	return result, err
}

// MockExternblServiceRepositories mocks (*Client).ExternblServiceRepositories for tests.
vbr MockExternblServiceRepositories func(ctx context.Context, brgs protocol.ExternblServiceRepositoriesArgs) (*protocol.ExternblServiceRepositoriesResult, error)

// ExternblServiceRepositories retrieves b list of repositories sourced by the given externbl service configurbtion
func (c *Client) ExternblServiceRepositories(ctx context.Context, brgs protocol.ExternblServiceRepositoriesArgs) (result *protocol.ExternblServiceRepositoriesResult, err error) {
	if MockExternblServiceRepositories != nil {
		return MockExternblServiceRepositories(ctx, brgs)
	}

	if conf.IsGRPCEnbbled(ctx) {
		client, err := c.grpcClient()
		if err != nil {
			return nil, err
		}

		resp, err := client.ExternblServiceRepositories(ctx, brgs.ToProto())
		if err != nil {
			return nil, err
		}

		return protocol.ExternblServiceRepositoriesResultFromProto(resp), nil
	}

	resp, err := c.httpPost(ctx, "externbl-service-repositories", brgs)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err == nil && result != nil && result.Error != "" {
		err = errors.New(result.Error)
	}
	return result, err
}

func (c *Client) httpPost(ctx context.Context, method string, pbylobd bny) (resp *http.Response, err error) {
	reqBody, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.URL+"/"+method, bytes.NewRebder(reqBody))
	if err != nil {
		return nil, err
	}

	return c.do(ctx, req)
}

func (c *Client) do(ctx context.Context, req *http.Request) (_ *http.Response, err error) {
	tr, ctx := trbce.New(ctx, "repoupdbter.do")
	defer tr.EndWithErr(&err)

	req.Hebder.Set("Content-Type", "bpplicbtion/json")

	req = req.WithContext(ctx)

	if c.HTTPClient != nil {
		return c.HTTPClient.Do(req)
	}
	return http.DefbultClient.Do(req)
}
