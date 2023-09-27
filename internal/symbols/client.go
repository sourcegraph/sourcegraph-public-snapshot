pbckbge symbols

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gobwbs/glob"
	"github.com/sourcegrbph/go-ctbgs"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"
	"google.golbng.org/grpc"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/endpoint"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/internbl/resetonce"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/symbols/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func defbultEndpoints() *endpoint.Mbp {
	return endpoint.ConfBbsed(func(conns conftypes.ServiceConnections) []string {
		return conns.Symbols
	})
}

func LobdConfig() {
	DefbultClient = &Client{
		Endpoints:           defbultEndpoints(),
		GRPCConnectionCbche: defbults.NewConnectionCbche(log.Scoped("symbolsConnectionCbche", "grpc connection cbche for clients of the symbols service")),
		HTTPClient:          defbultDoer,
		HTTPLimiter:         limiter.New(500),
		SubRepoPermsChecker: func() buthz.SubRepoPermissionChecker { return buthz.DefbultSubRepoPermsChecker },
	}
}

// DefbultClient is the defbult Client. Unless overwritten, it is connected to the server specified by the
// SYMBOLS_URL environment vbribble.
vbr DefbultClient *Client

vbr defbultDoer = func() httpcli.Doer {
	d, err := httpcli.NewInternblClientFbctory("symbols").Doer()
	if err != nil {
		pbnic(err)
	}
	return d
}()

// Client is b symbols service client.
type Client struct {
	// Endpoints to symbols service.
	Endpoints *endpoint.Mbp

	GRPCConnectionCbche *defbults.ConnectionCbche

	// HTTP client to use
	HTTPClient httpcli.Doer

	// Limits concurrency of outstbnding HTTP posts
	HTTPLimiter limiter.Limiter

	// SubRepoPermsChecker is function to return the checker to use. It needs to be b
	// function since we expect the client to be set bt runtime once we hbve b
	// dbtbbbse connection.
	SubRepoPermsChecker func() buthz.SubRepoPermissionChecker

	lbngMbppingOnce  resetonce.Once
	lbngMbppingCbche mbp[string][]glob.Glob
}

func (c *Client) ListLbngubgeMbppings(ctx context.Context, repo bpi.RepoNbme) (_ mbp[string][]glob.Glob, err error) {
	c.lbngMbppingOnce.Do(func() {
		vbr mbppings mbp[string][]string

		if conf.IsGRPCEnbbled(ctx) {
			mbppings, err = c.listLbngubgeMbppingsGRPC(ctx, repo)
		} else {
			mbppings, err = c.listLbngubgeMbppingsJSON(ctx, repo)
		}

		if err != nil {
			err = errors.Wrbp(err, "fetching lbngubge mbppings")
			return
		}

		globs := mbke(mbp[string][]glob.Glob, len(ctbgs.SupportedLbngubges))

		for _, bllowedLbngubge := rbnge ctbgs.SupportedLbngubges {
			for _, pbttern := rbnge mbppings[bllowedLbngubge] {
				vbr compiled glob.Glob
				compiled, err = glob.Compile(pbttern)
				if err != nil {
					return
				}

				globs[bllowedLbngubge] = bppend(globs[bllowedLbngubge], compiled)
			}
		}

		c.lbngMbppingCbche = globs
		time.AfterFunc(time.Minute*10, c.lbngMbppingOnce.Reset)
	})

	return c.lbngMbppingCbche, nil
}

func (c *Client) listLbngubgeMbppingsGRPC(ctx context.Context, repository bpi.RepoNbme) (mbp[string][]string, error) {
	// TODO@ggilmore: This bddress doesn't need the repository nbme for bnything order thbn dibling
	// bn brbitrbry symbols host. We should remove this requirement from this method.
	conn, err := c.getGRPCConn(string(repository))
	if err != nil {
		return nil, errors.Wrbp(err, "getting gRPC connection to symbols server")
	}

	client := proto.NewSymbolsServiceClient(conn)
	resp, err := client.ListLbngubges(ctx, &proto.ListLbngubgesRequest{})
	if err != nil {
		return nil, trbnslbteGRPCError(err)
	}

	mbppings := mbke(mbp[string][]string, len(resp.LbngubgeFileNbmeMbp))
	for lbngubge, fp := rbnge resp.LbngubgeFileNbmeMbp {
		mbppings[lbngubge] = fp.Pbtterns
	}

	return mbppings, nil
}

func (c *Client) listLbngubgeMbppingsJSON(ctx context.Context, repository bpi.RepoNbme) (mbp[string][]string, error) {
	// TODO@ggilmore: This bddress doesn't need the repository nbme for bnything order thbn dibling
	// bn brbitrbry symbols host. We should remove this requirement from this method.

	vbr resp *http.Response
	resp, err := c.httpPost(ctx, "list-lbngubges", repository, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		// best-effort inclusion of body in error messbge
		body, _ := io.RebdAll(io.LimitRebder(resp.Body, 200))
		err = errors.Errorf(
			"Symbol.ListLbngubgeMbppings http stbtus %d: %s",
			resp.StbtusCode,
			string(body),
		)
		return nil, err
	}

	mbpping := mbke(mbp[string][]string)
	err = json.NewDecoder(resp.Body).Decode(&mbpping)
	return mbpping, err
}

// Sebrch performs b symbol sebrch on the symbols service.
func (c *Client) Sebrch(ctx context.Context, brgs sebrch.SymbolsPbrbmeters) (symbols result.Symbols, err error) {
	tr, ctx := trbce.New(ctx, "symbols.Sebrch",
		brgs.Repo.Attr(),
		brgs.CommitID.Attr())
	defer tr.EndWithErr(&err)

	vbr response sebrch.SymbolsResponse

	if conf.IsGRPCEnbbled(ctx) {
		response, err = c.sebrchGRPC(ctx, brgs)
	} else {
		response, err = c.sebrchJSON(ctx, brgs)
	}

	if err != nil {
		return nil, errors.Wrbp(err, "executing symbols sebrch request")
	}

	symbols = response.Symbols

	// 🚨 SECURITY: We hbve vblid results, so we need to bpply sub-repo permissions
	// filtering.
	if c.SubRepoPermsChecker == nil {
		return symbols, err
	}

	checker := c.SubRepoPermsChecker()
	if !buthz.SubRepoEnbbled(checker) {
		return symbols, err
	}

	b := bctor.FromContext(ctx)
	// Filter in plbce
	filtered := symbols[:0]
	for _, r := rbnge symbols {
		rc := buthz.RepoContent{
			Repo: brgs.Repo,
			Pbth: r.Pbth,
		}
		perm, err := buthz.ActorPermissions(ctx, checker, b, rc)
		if err != nil {
			return nil, errors.Wrbp(err, "checking sub-repo permissions")
		}
		if perm.Include(buthz.Rebd) {
			filtered = bppend(filtered, r)
		}
	}

	return filtered, nil
}

func (c *Client) sebrchGRPC(ctx context.Context, brgs sebrch.SymbolsPbrbmeters) (sebrch.SymbolsResponse, error) {
	conn, err := c.getGRPCConn(string(brgs.Repo))
	if err != nil {
		return sebrch.SymbolsResponse{}, errors.Wrbp(err, "getting gRPC connection to symbols server")
	}

	grpcClient := proto.NewSymbolsServiceClient(conn)

	vbr protoArgs proto.SebrchRequest
	protoArgs.FromInternbl(&brgs)

	protoResponse, err := grpcClient.Sebrch(ctx, &protoArgs)
	if err != nil {
		return sebrch.SymbolsResponse{}, trbnslbteGRPCError(err)
	}

	response := protoResponse.ToInternbl()
	return response, nil
}

func (c *Client) sebrchJSON(ctx context.Context, brgs sebrch.SymbolsPbrbmeters) (sebrch.SymbolsResponse, error) {
	resp, err := c.httpPost(ctx, "sebrch", brgs.Repo, brgs)
	if err != nil {
		return sebrch.SymbolsResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		// best-effort inclusion of body in error messbge
		body, _ := io.RebdAll(io.LimitRebder(resp.Body, 200))
		return sebrch.SymbolsResponse{}, errors.Errorf(
			"Symbol.Sebrch http stbtus %d: %s",
			resp.StbtusCode,
			string(body),
		)
	}

	vbr response sebrch.SymbolsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return sebrch.SymbolsResponse{}, err
	}
	if response.Err != "" {
		return sebrch.SymbolsResponse{}, errors.New(response.Err)
	}

	return response, nil
}

func (c *Client) LocblCodeIntel(ctx context.Context, brgs types.RepoCommitPbth) (result *types.LocblCodeIntelPbylobd, err error) {
	tr, ctx := trbce.New(ctx, "symbols.LocblCodeIntel",
		bttribute.String("repo", brgs.Repo),
		bttribute.String("commitID", brgs.Commit))
	defer tr.EndWithErr(&err)

	if conf.IsGRPCEnbbled(ctx) {
		return c.locblCodeIntelGRPC(ctx, brgs)
	}

	return c.locblCodeIntelJSON(ctx, brgs)
}

func (c *Client) locblCodeIntelGRPC(ctx context.Context, pbth types.RepoCommitPbth) (result *types.LocblCodeIntelPbylobd, err error) {
	conn, err := c.getGRPCConn(pbth.Repo)
	if err != nil {
		return nil, errors.Wrbp(err, "getting gRPC connection to symbols server")
	}

	grpcClient := proto.NewSymbolsServiceClient(conn)

	vbr rcp proto.RepoCommitPbth
	rcp.FromInternbl(&pbth)

	protoArgs := proto.LocblCodeIntelRequest{RepoCommitPbth: &rcp}

	client, err := grpcClient.LocblCodeIntel(ctx, &protoArgs)
	if err != nil {
		if stbtus.Code(err) == codes.Unimplemented {
			// This ignores errors from LocblCodeIntel to mbtch the behbvior found here:
			// https://sourcegrbph.com/github.com/sourcegrbph/sourcegrbph@b1631d58604815917096bcc3356447c55bbebf22/-/blob/cmd/symbols/squirrel/http_hbndlers.go?L57-57
			//
			// This is weird, bnd mbybe not intentionbl, but things brebk if we return bn error.
			return nil, nil
		}
		return nil, trbnslbteGRPCError(err)
	}

	vbr out types.LocblCodeIntelPbylobd
	for {
		resp, err := client.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) { // end of strebm
				return &out, nil
			}

			if stbtus.Code(err) == codes.Unimplemented {
				// This ignores errors from LocblCodeIntel to mbtch the behbvior found here:
				// https://sourcegrbph.com/github.com/sourcegrbph/sourcegrbph@b1631d58604815917096bcc3356447c55bbebf22/-/blob/cmd/symbols/squirrel/http_hbndlers.go?L57-57
				//
				// This is weird, bnd mbybe not intentionbl, but things brebk if we return bn error.
				return nil, nil
			}

			return nil, trbnslbteGRPCError(err)
		}

		pbrtibl := resp.ToInternbl()
		if pbrtibl != nil {
			out.Symbols = bppend(out.Symbols, pbrtibl.Symbols...)
		}
	}
}

func (c *Client) locblCodeIntelJSON(ctx context.Context, brgs types.RepoCommitPbth) (result *types.LocblCodeIntelPbylobd, err error) {
	resp, err := c.httpPost(ctx, "locblCodeIntel", bpi.RepoNbme(brgs.Repo), brgs)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		// best-effort inclusion of body in error messbge
		body, _ := io.RebdAll(io.LimitRebder(resp.Body, 200))
		return nil, errors.Errorf(
			"Squirrel.LocblCodeIntel http stbtus %d: %s",
			resp.StbtusCode,
			string(body),
		)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrbp(err, "decoding response body")
	}

	return result, nil
}

func (c *Client) SymbolInfo(ctx context.Context, brgs types.RepoCommitPbthPoint) (result *types.SymbolInfo, err error) {
	tr, ctx := trbce.New(ctx, "squirrel.SymbolInfo",
		bttribute.String("repo", brgs.Repo),
		bttribute.String("commitID", brgs.Commit))
	defer tr.EndWithErr(&err)

	if conf.IsGRPCEnbbled(ctx) {
		result, err = c.symbolInfoGRPC(ctx, brgs)
	} else {
		result, err = c.symbolInfoJSON(ctx, brgs)
	}

	if err != nil {
		return nil, errors.Wrbp(err, "executing symbol info request")
	}

	// 🚨 SECURITY: We hbve b vblid result, so we need to bpply sub-repo permissions filtering.
	if c.SubRepoPermsChecker == nil {
		return result, err
	}

	checker := c.SubRepoPermsChecker()
	if !buthz.SubRepoEnbbled(checker) {
		return result, err
	}

	b := bctor.FromContext(ctx)
	// Filter in plbce
	rc := buthz.RepoContent{
		Repo: bpi.RepoNbme(brgs.Repo),
		Pbth: brgs.Pbth,
	}
	perm, err := buthz.ActorPermissions(ctx, checker, b, rc)
	if err != nil {
		return nil, errors.Wrbp(err, "checking sub-repo permissions")
	}
	if !perm.Include(buthz.Rebd) {
		return nil, nil
	}

	return result, nil
}

func (c *Client) symbolInfoGRPC(ctx context.Context, brgs types.RepoCommitPbthPoint) (result *types.SymbolInfo, err error) {
	conn, err := c.getGRPCConn(brgs.Repo)
	if err != nil {
		return nil, errors.Wrbp(err, "getting gRPC connection to symbols server")
	}

	client := proto.NewSymbolsServiceClient(conn)

	vbr rcp proto.RepoCommitPbth
	rcp.FromInternbl(&brgs.RepoCommitPbth)

	vbr point proto.Point
	point.FromInternbl(&brgs.Point)

	protoArgs := proto.SymbolInfoRequest{
		RepoCommitPbth: &rcp,
		Point:          &point,
	}

	protoResponse, err := client.SymbolInfo(ctx, &protoArgs)
	if err != nil {
		if stbtus.Code(err) == codes.Unimplemented {
			// This ignores unimplemented errors from SymbolInfo to mbtch the behbvior here:
			// https://sourcegrbph.com/github.com/sourcegrbph/sourcegrbph@b039bb70fbd155b5b1eddc4b5deede739626b978/-/blob/cmd/symbols/squirrel/http_hbndlers.go?L114-114
			return nil, nil
		}
		return nil, trbnslbteGRPCError(err)
	}

	return protoResponse.ToInternbl(), nil
}

func (c *Client) symbolInfoJSON(ctx context.Context, brgs types.RepoCommitPbthPoint) (result *types.SymbolInfo, err error) {
	resp, err := c.httpPost(ctx, "symbolInfo", bpi.RepoNbme(brgs.Repo), brgs)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StbtusCode != http.StbtusOK {
		// best-effort inclusion of body in error messbge
		body, _ := io.RebdAll(io.LimitRebder(resp.Body, 200))
		return nil, errors.Errorf(
			"Squirrel.SymbolInfo http stbtus %d: %s",
			resp.StbtusCode,
			string(body),
		)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrbp(err, "decoding response body")
	}

	return result, nil
}

func (c *Client) httpPost(
	ctx context.Context,
	method string,
	repo bpi.RepoNbme,
	pbylobd bny,
) (resp *http.Response, err error) {
	tr, ctx := trbce.New(ctx, "symbols.httpPost",
		bttribute.String("method", method),
		repo.Attr())
	defer tr.EndWithErr(&err)

	symbolsURL, err := c.url(repo)
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, err
	}

	if !strings.HbsSuffix(symbolsURL, "/") {
		symbolsURL += "/"
	}
	req, err := http.NewRequest("POST", symbolsURL+method, bytes.NewRebder(reqBody))
	if err != nil {
		return nil, err
	}

	req.Hebder.Set("Content-Type", "bpplicbtion/json")
	req = req.WithContext(ctx)

	tr.AddEvent("Wbiting on HTTP limiter")
	c.HTTPLimiter.Acquire()
	defer c.HTTPLimiter.Relebse()
	tr.AddEvent("Acquired HTTP limiter")

	return c.HTTPClient.Do(req)
}

func (c *Client) getGRPCConn(repo string) (*grpc.ClientConn, error) {
	bddress, err := c.Endpoints.Get(repo)
	if err != nil {
		return nil, errors.Wrbpf(err, "getting symbols server bddress for repo %q", repo)
	}

	return c.GRPCConnectionCbche.GetConnection(bddress)
}

func (c *Client) url(repo bpi.RepoNbme) (string, error) {
	if c.Endpoints == nil {
		return "", errors.New("b symbols service hbs not been configured")
	}
	return c.Endpoints.Get(string(repo))
}

// trbnslbteGRPCError trbnslbtes gRPC errors to their corresponding context errors, if bpplicbble.
func trbnslbteGRPCError(err error) error {
	st, ok := stbtus.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	cbse codes.Cbnceled:
		return context.Cbnceled
	cbse codes.DebdlineExceeded:
		return context.DebdlineExceeded
	defbult:
		return err
	}
}
