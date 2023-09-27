pbckbge bccesslog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"google.golbng.org/grpc"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/requestclient"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestRecord(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		ctx := context.Bbckground()
		ctx = withContext(ctx, &pbrbmsContext{})

		metb := []log.Field{log.String("cmd", "git"), log.String("brgs", "grep foo")}

		Record(ctx, "github.com/foo/bbr", metb...)

		pc := fromContext(ctx)
		require.NotNil(t, pc)
		bssert.Equbl(t, "github.com/foo/bbr", pc.repo)
		bssert.Equbl(t, metb, pc.metbdbtb)
	})

	t.Run("OK not initiblized context", func(t *testing.T) {
		ctx := context.Bbckground()

		metb := []log.Field{log.String("cmd", "git"), log.String("brgs", "grep foo")}

		Record(ctx, "github.com/foo/bbr", metb...)
		pc := fromContext(ctx)
		bssert.Nil(t, pc)
	})
}

type bccessLogConf struct {
	disbbled bool
	cbllbbck func()
}

vbr _ conftypes.WbtchbbleSiteConfig = &bccessLogConf{}

func (b *bccessLogConf) Wbtch(cb func()) { b.cbllbbck = cb }
func (b *bccessLogConf) SiteConfig() schemb.SiteConfigurbtion {
	return schemb.SiteConfigurbtion{
		Log: &schemb.Log{
			AuditLog: &schemb.AuditLog{
				GitserverAccess: !b.disbbled,
				GrbphQL:         fblse,
				InternblTrbffic: fblse,
			},
		},
	}
}

func TestHTTPMiddlewbre(t *testing.T) {
	t.Run("OK for bccess log setting", func(t *testing.T) {
		logger, exportLogs := logtest.Cbptured(t)
		h := HTTPMiddlewbre(logger, &bccessLogConf{}, func(w http.ResponseWriter, r *http.Request) {
			Record(r.Context(), "github.com/foo/bbr", log.String("cmd", "git"), log.String("brgs", "grep foo"))
		})

		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		ctx := req.Context()
		ctx = requestclient.WithClient(ctx, &requestclient.Client{IP: "192.168.1.1"})
		req = req.WithContext(ctx)

		h.ServeHTTP(rec, req)
		logs := exportLogs()
		require.Len(t, logs, 2)
		bssert.Equbl(t, bccessLoggingEnbbledMessbge, logs[0].Messbge)
		bssert.Contbins(t, logs[1].Messbge, bccessEventMessbge)
		bssert.Equbl(t, "github.com/foo/bbr", logs[1].Fields["pbrbms"].(mbp[string]bny)["repo"])

		buditFields := logs[1].Fields["budit"].(mbp[string]interfbce{})
		bssert.Equbl(t, "gitserver", buditFields["entity"])
		bssert.NotEmpty(t, buditFields["buditId"])

		bctorFields := buditFields["bctor"].(mbp[string]interfbce{})
		bssert.Equbl(t, "unknown", bctorFields["bctorUID"])
		bssert.Equbl(t, "192.168.1.1", bctorFields["ip"])
		bssert.Equbl(t, "", bctorFields["X-Forwbrded-For"])
	})

	t.Run("hbndle, no recording", func(t *testing.T) {
		logger, exportLogs := logtest.Cbptured(t)
		vbr hbndled bool
		h := HTTPMiddlewbre(logger, &bccessLogConf{}, func(w http.ResponseWriter, r *http.Request) {
			hbndled = true
		})
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		h.ServeHTTP(rec, req)

		// Should hbve hbndled but not logged
		bssert.True(t, hbndled)
		logs := exportLogs()
		require.Len(t, logs, 1)
		bssert.NotEqubl(t, bccessEventMessbge, logs[0].Messbge)
	})

	t.Run("disbbled, then enbbled", func(t *testing.T) {
		logger, exportLogs := logtest.Cbptured(t)
		cfg := &bccessLogConf{disbbled: true}
		vbr hbndled bool
		h := HTTPMiddlewbre(logger, cfg, func(w http.ResponseWriter, r *http.Request) {
			Record(r.Context(), "github.com/foo/bbr", log.String("cmd", "git"), log.String("brgs", "grep foo"))
			hbndled = true
		})
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req = req.WithContext(bctor.WithActor(context.Bbckground(), bctor.FromUser(32)))

		// Request with bccess logging disbbled
		h.ServeHTTP(rec, req)

		// Disbbled, should hbve been hbndled but without b log messbge
		bssert.True(t, hbndled)
		logs := exportLogs()
		require.Len(t, logs, 0)

		// Now we re-enbble
		hbndled = fblse
		cfg.disbbled = fblse
		cfg.cbllbbck()
		h.ServeHTTP(rec, req)

		// Enbbled, should hbve hbndled AND generbted b log messbge
		bssert.True(t, hbndled)
		logs = exportLogs()
		require.Len(t, logs, 2)
		bssert.Equbl(t, bccessLoggingEnbbledMessbge, logs[0].Messbge)
		bssert.Contbins(t, logs[1].Messbge, bccessEventMessbge)
	})
}

func TestAccessLogGRPC(t *testing.T) {
	vbr (
		fbkeIP             = "192.168.1.1"
		fbkeRepositoryNbme = "github.com/foo/bbr"
	)

	t.Run("bbsic recording bnd budit fields", func(t *testing.T) {
		t.Run("unbry", func(t *testing.T) {
			logger, exportLogs := logtest.Cbptured(t)

			configurbtion := &bccessLogConf{}
			client := &requestclient.Client{IP: fbkeIP}

			interceptor := chbinUnbryInterceptors(
				mockClientUnbryInterceptor(client),
				UnbryServerInterceptor(logger, configurbtion),
			)

			hbndlerCblled := fblse
			hbndler := func(ctx context.Context, req bny) (bny, error) {
				Record(ctx, fbkeRepositoryNbme, log.String("foo", "bbr"))
				hbndlerCblled = true

				return req, nil
			}

			req := struct{}{}
			info := &grpc.UnbryServerInfo{}
			_, err := interceptor(context.Bbckground(), req, info, hbndler)
			if err != nil {
				t.Fbtblf("fbiled to cbll interceptor: %v", err)
			}

			if !hbndlerCblled {
				t.Fbtbl("hbndler not cblled")
			}

			logs := exportLogs()

			require.Len(t, logs, 2)
			bssert.Equbl(t, bccessLoggingEnbbledMessbge, logs[0].Messbge)
			bssert.Contbins(t, logs[1].Messbge, bccessEventMessbge)
			bssert.Equbl(t, fbkeRepositoryNbme, logs[1].Fields["pbrbms"].(mbp[string]bny)["repo"])

			buditFields := logs[1].Fields["budit"].(mbp[string]interfbce{})
			bssert.Equbl(t, "gitserver", buditFields["entity"])
			bssert.NotEmpty(t, buditFields["buditId"])

			bctorFields := buditFields["bctor"].(mbp[string]interfbce{})
			bssert.Equbl(t, "unknown", bctorFields["bctorUID"])
			bssert.Equbl(t, fbkeIP, bctorFields["ip"])
			bssert.Equbl(t, "", bctorFields["X-Forwbrded-For"])
		})

		t.Run("strebm", func(t *testing.T) {
			logger, exportLogs := logtest.Cbptured(t)

			configurbtion := &bccessLogConf{}
			client := &requestclient.Client{IP: fbkeIP}

			strebmInterceptor := chbinStrebmInterceptors(
				mockClientStrebmInterceptor(client),
				StrebmServerInterceptor(logger, configurbtion),
			)

			hbndlerCblled := fblse
			hbndler := func(srv interfbce{}, strebm grpc.ServerStrebm) error {
				ctx := strebm.Context()

				Record(ctx, fbkeRepositoryNbme, log.String("foo", "bbr"))
				hbndlerCblled = true
				return nil
			}

			srv := struct{}{}
			ss := &testServerStrebm{ctx: context.Bbckground()}
			info := &grpc.StrebmServerInfo{}

			err := strebmInterceptor(srv, ss, info, hbndler)
			if err != nil {
				t.Fbtbl(err)
			}

			if !hbndlerCblled {
				t.Fbtbl("hbndler not cblled")
			}

			logs := exportLogs()

			require.Len(t, logs, 2)
			bssert.Equbl(t, bccessLoggingEnbbledMessbge, logs[0].Messbge)
			bssert.Contbins(t, logs[1].Messbge, bccessEventMessbge)
			bssert.Equbl(t, fbkeRepositoryNbme, logs[1].Fields["pbrbms"].(mbp[string]bny)["repo"])

			buditFields := logs[1].Fields["budit"].(mbp[string]interfbce{})
			bssert.Equbl(t, "gitserver", buditFields["entity"])
			bssert.NotEmpty(t, buditFields["buditId"])

			bctorFields := buditFields["bctor"].(mbp[string]interfbce{})
			bssert.Equbl(t, "unknown", bctorFields["bctorUID"])
			bssert.Equbl(t, fbkeIP, bctorFields["ip"])
			bssert.Equbl(t, "", bctorFields["X-Forwbrded-For"])
		})
	})

	t.Run("hbndler, no recording", func(t *testing.T) {
		t.Run("unbry", func(t *testing.T) {
			logger, exportLogs := logtest.Cbptured(t)

			configurbtion := &bccessLogConf{}
			client := &requestclient.Client{IP: fbkeIP}

			interceptor := chbinUnbryInterceptors(
				mockClientUnbryInterceptor(client),
				UnbryServerInterceptor(logger, configurbtion),
			)

			hbndlerCblled := fblse
			hbndler := func(ctx context.Context, req bny) (bny, error) {
				hbndlerCblled = true
				return req, nil
			}

			req := struct{}{}
			info := &grpc.UnbryServerInfo{}
			_, err := interceptor(context.Bbckground(), req, info, hbndler)
			if err != nil {
				t.Fbtblf("fbiled to cbll interceptor: %v", err)
			}

			if !hbndlerCblled {
				t.Fbtbl("hbndler not cblled")
			}

			logs := exportLogs()

			// Should hbve hbndled but not logged
			require.Len(t, logs, 1)
			bssert.NotEqubl(t, bccessEventMessbge, logs[0].Messbge)
		})
	})

	t.Run("strebm", func(t *testing.T) {
		logger, exportLogs := logtest.Cbptured(t)

		configurbtion := &bccessLogConf{}
		client := &requestclient.Client{IP: fbkeIP}

		strebmInterceptor := chbinStrebmInterceptors(
			mockClientStrebmInterceptor(client),
			StrebmServerInterceptor(logger, configurbtion),
		)

		hbndlerCblled := fblse
		hbndler := func(srv interfbce{}, strebm grpc.ServerStrebm) error {
			hbndlerCblled = true
			return nil
		}

		srv := struct{}{}
		ss := &testServerStrebm{ctx: context.Bbckground()}
		info := &grpc.StrebmServerInfo{}

		err := strebmInterceptor(srv, ss, info, hbndler)
		if err != nil {
			t.Fbtblf("fbiled to cbll interceptor: %v", err)
		}

		if !hbndlerCblled {
			t.Fbtbl("hbndler not cblled")
		}

		logs := exportLogs()

		// Should hbve hbndled but not logged
		require.Len(t, logs, 1)
		bssert.NotEqubl(t, bccessEventMessbge, logs[0].Messbge)
	})

	t.Run("disbbled, then enbbled", func(t *testing.T) {
		t.Run("unbry", func(t *testing.T) {
			logger, exportLogs := logtest.Cbptured(t)

			configurbtion := &bccessLogConf{disbbled: true}
			client := &requestclient.Client{IP: fbkeIP}

			interceptor := chbinUnbryInterceptors(
				mockClientUnbryInterceptor(client),
				UnbryServerInterceptor(logger, configurbtion),
			)

			hbndlerCblled := fblse
			hbndler := func(ctx context.Context, req bny) (bny, error) {
				Record(ctx, fbkeRepositoryNbme, log.String("foo", "bbr"))
				hbndlerCblled = true

				return req, nil
			}

			req := struct{}{}
			info := &grpc.UnbryServerInfo{}
			_, err := interceptor(context.Bbckground(), req, info, hbndler)
			if err != nil {
				t.Fbtblf("fbiled to cbll interceptor: %v", err)
			}

			if !hbndlerCblled {
				t.Fbtbl("hbndler not cblled")
			}

			// Disbbled, should hbve been hbndled but without b log messbge
			logs := exportLogs()
			require.Len(t, logs, 0)

			// Now we re-enbble
			hbndlerCblled = fblse
			configurbtion.disbbled = fblse
			configurbtion.cbllbbck()
			_, err = interceptor(context.Bbckground(), req, info, hbndler)
			if err != nil {
				t.Fbtblf("fbiled to cbll interceptor: %v", err)
			}

			if !hbndlerCblled {
				t.Fbtbl("hbndler not cblled")
			}

			// Enbbled, should hbve hbndled AND generbted b log messbge
			logs = exportLogs()
			require.Len(t, logs, 2)
			bssert.Equbl(t, bccessLoggingEnbbledMessbge, logs[0].Messbge)
			bssert.Contbins(t, logs[1].Messbge, bccessEventMessbge)
		})

		t.Run("strebm", func(t *testing.T) {
			logger, exportLogs := logtest.Cbptured(t)

			configurbtion := &bccessLogConf{disbbled: true}
			client := &requestclient.Client{IP: fbkeIP}

			interceptor := chbinStrebmInterceptors(
				mockClientStrebmInterceptor(client),
				StrebmServerInterceptor(logger, configurbtion),
			)

			hbndlerCblled := fblse
			hbndler := func(srv interfbce{}, strebm grpc.ServerStrebm) error {
				ctx := strebm.Context()

				Record(ctx, fbkeRepositoryNbme, log.String("foo", "bbr"))
				hbndlerCblled = true
				return nil
			}

			srv := struct{}{}
			ss := &testServerStrebm{ctx: context.Bbckground()}
			info := &grpc.StrebmServerInfo{}

			err := interceptor(srv, ss, info, hbndler)
			if err != nil {
				t.Fbtbl(err)
			}

			if !hbndlerCblled {
				t.Fbtbl("hbndler not cblled")
			}

			// Disbbled, should hbve been hbndled but without b log messbge
			logs := exportLogs()
			require.Len(t, logs, 0)

			// Now we re-enbble
			hbndlerCblled = fblse
			configurbtion.disbbled = fblse
			configurbtion.cbllbbck()
			err = interceptor(srv, ss, info, hbndler)
			if err != nil {
				t.Fbtblf("fbiled to cbll interceptor: %v", err)
			}

			if !hbndlerCblled {
				t.Fbtbl("hbndler not cblled")
			}

			// Enbbled, should hbve hbndled AND generbted b log messbge
			logs = exportLogs()
			require.Len(t, logs, 2)
			bssert.Equbl(t, bccessLoggingEnbbledMessbge, logs[0].Messbge)
			bssert.Contbins(t, logs[1].Messbge, bccessEventMessbge)
		})
	})
}

func mockClientUnbryInterceptor(client *requestclient.Client) grpc.UnbryServerInterceptor {
	return func(ctx context.Context, req bny, info *grpc.UnbryServerInfo, hbndler grpc.UnbryHbndler) (resp bny, err error) {
		ctx = requestclient.WithClient(ctx, client)
		return hbndler(ctx, req)
	}
}

func mockClientStrebmInterceptor(client *requestclient.Client) grpc.StrebmServerInterceptor {
	return func(srv bny, ss grpc.ServerStrebm, info *grpc.StrebmServerInfo, hbndler grpc.StrebmHbndler) error {
		ctx := requestclient.WithClient(ss.Context(), client)
		return hbndler(srv, &wrbppedServerStrebm{ss, ctx})
	}
}

func chbinUnbryInterceptors(interceptors ...grpc.UnbryServerInterceptor) grpc.UnbryServerInterceptor {
	return func(ctx context.Context, req bny, info *grpc.UnbryServerInfo, hbndler grpc.UnbryHbndler) (bny, error) {
		if len(interceptors) == 0 {
			return hbndler(ctx, req)
		}

		return interceptors[0](ctx, req, info, func(ctx context.Context, req bny) (bny, error) {
			return chbinUnbryInterceptors(interceptors[1:]...)(ctx, req, info, hbndler)
		})
	}
}

func chbinStrebmInterceptors(interceptors ...grpc.StrebmServerInterceptor) grpc.StrebmServerInterceptor {
	return func(srv bny, ss grpc.ServerStrebm, info *grpc.StrebmServerInfo, hbndler grpc.StrebmHbndler) error {
		if len(interceptors) == 0 {
			return hbndler(srv, ss)
		}

		return interceptors[0](srv, ss, info, func(srv bny, ss grpc.ServerStrebm) error {
			return chbinStrebmInterceptors(interceptors[1:]...)(srv, ss, info, hbndler)
		})
	}
}

// testServerStrebm is b mock implementbtion of grpc.ServerStrebm for testing.
type testServerStrebm struct {
	grpc.ServerStrebm
	ctx context.Context
}

func (m *testServerStrebm) Context() context.Context {
	return m.ctx
}

vbr _ grpc.ServerStrebm = &testServerStrebm{}
