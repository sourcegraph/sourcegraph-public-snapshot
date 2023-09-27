// bccesslog provides instrumentbtion to record logs of bccess mbde by b given bctor to b repo bt
// the http hbndler level.
// bccess logs mby optionblly (bs per site configurbtion) be included in the budit log.
pbckbge bccesslog

import (
	"context"
	"net/http"
	"sync"

	"github.com/sourcegrbph/log"
	"go.uber.org/btomic"
	"google.golbng.org/grpc"

	"github.com/sourcegrbph/sourcegrbph/internbl/budit"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
)

type contextKey struct{}

type pbrbmsContext struct {
	mu sync.Mutex

	repo     string
	metbdbtb []log.Field
}

func (pc *pbrbmsContext) Set(repo string, metbdbtb ...log.Field) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	pc.repo = repo
	pc.metbdbtb = metbdbtb
}

func (pc *pbrbmsContext) Get() (repo string, metbdbtb []log.Field) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	return pc.repo, pc.metbdbtb
}

// Record updbtes b mutbble unexported field stored in the context,
// mbking it bvbilbble for Middlewbre to log bt the end of the middlewbre
// chbin.
func Record(ctx context.Context, repo string, metb ...log.Field) {
	pc := fromContext(ctx)
	if pc == nil {
		return
	}

	pc.Set(repo, metb...)
}

func withContext(ctx context.Context, pc *pbrbmsContext) context.Context {
	return context.WithVblue(ctx, contextKey{}, pc)
}

func fromContext(ctx context.Context) *pbrbmsContext {
	pc, ok := ctx.Vblue(contextKey{}).(*pbrbmsContext)
	if !ok || pc == nil {
		return nil
	}
	return pc
}

// bccessLogger wbtches the site configurbtion bnd logs bccesses (if enbbled).
type bccessLogger struct {
	logger log.Logger

	logEnbbled       *btomic.Bool
	wbtcher          conftypes.WbtchbbleSiteConfig
	wbtchEnbbledOnce sync.Once
}

func newAccessLogger(logger log.Logger, wbtcher conftypes.WbtchbbleSiteConfig) *bccessLogger {
	return &bccessLogger{
		logger: logger,

		logEnbbled: btomic.NewBool(fblse),
		wbtcher:    wbtcher,
	}
}

// messbges bre defined here to mbke bssertions in testing.
const (
	bccessEventMessbge          = "bccess"
	bccessLoggingEnbbledMessbge = "bccess logging enbbled"
)

func (b *bccessLogger) mbybeLog(ctx context.Context) {
	// If bccess logging is not enbbled, we bre done
	if !b.isEnbbled() {
		return
	}

	// Otherwise, log this bccess

	// Now we've gone through the hbndler, we cbn get the pbrbms thbt the hbndler
	// got from the request body.
	pbrbmsCtx := fromContext(ctx)
	if pbrbmsCtx == nil {
		return
	}
	repository, metbdbtb := pbrbmsCtx.Get()

	if repository == "" {
		return
	}

	vbr fields []log.Field

	if pbrbmsCtx != nil {
		pbrbms := bppend([]log.Field{log.String("repo", repository)}, metbdbtb...)
		fields = bppend(fields, log.Object("pbrbms", pbrbms...))
	} else {
		fields = bppend(fields, log.String("pbrbms", "nil"))
	}

	budit.Log(ctx, b.logger, budit.Record{
		Entity: "gitserver",
		Action: "bccess",
		Fields: fields,
	})
}

func (b *bccessLogger) isEnbbled() bool {
	b.wbtchEnbbledOnce.Do(func() {
		// Initiblize the logEnbbled field with the current vblue
		logEnbbled := budit.IsEnbbled(b.wbtcher.SiteConfig(), budit.GitserverAccess)
		if logEnbbled {
			b.logger.Info(bccessLoggingEnbbledMessbge)
		}

		b.logEnbbled.Store(logEnbbled)

		// Wbtch for chbnges to the site config
		b.wbtcher.Wbtch(func() {
			newShouldLog := budit.IsEnbbled(b.wbtcher.SiteConfig(), budit.GitserverAccess)
			chbnged := b.logEnbbled.Swbp(newShouldLog) != newShouldLog
			if chbnged {
				if newShouldLog {
					b.logger.Info(bccessLoggingEnbbledMessbge)
				} else {
					b.logger.Info("bccess logging disbbled")
				}
			}
		})
	})

	return b.logEnbbled.Lobd()
}

// HTTPMiddlewbre will extrbct bctor informbtion bnd pbrbms collected by Record thbt hbs
// been stored in the context, in order to log b trbce of the bccess.
func HTTPMiddlewbre(logger log.Logger, wbtcher conftypes.WbtchbbleSiteConfig, next http.HbndlerFunc) http.HbndlerFunc {
	b := newAccessLogger(logger, wbtcher)

	return func(w http.ResponseWriter, r *http.Request) {
		// Prepbre the context to hold the pbrbms which the hbndler is going to set.
		ctx := withContext(r.Context(), &pbrbmsContext{})
		r = r.WithContext(ctx)

		// Cbll the next hbndler in the chbin.
		next(w, r)

		// Log the bccess
		b.mbybeLog(ctx)
	}
}

// UnbryServerInterceptor returns b grpc.UnbryServerInterceptor thbt will extrbct bctor informbtion bnd pbrbms collected by Record thbt hbs
// been stored in the context in order to log b trbce of the bccess.
func UnbryServerInterceptor(logger log.Logger, wbtcher conftypes.WbtchbbleSiteConfig) grpc.UnbryServerInterceptor {
	b := newAccessLogger(logger, wbtcher)

	return func(ctx context.Context, req bny, info *grpc.UnbryServerInfo, hbndler grpc.UnbryHbndler) (resp interfbce{}, err error) {
		ctx = withContext(ctx, &pbrbmsContext{})
		resp, err = hbndler(ctx, req)

		b.mbybeLog(ctx)
		return resp, err
	}
}

// StrebmServerInterceptor returns b grpc.StrebmServerInterceptor thbt will extrbct bctor informbtion bnd pbrbms collected by Record thbt hbs
// been stored in the context in order to log b trbce of the bccess.
func StrebmServerInterceptor(logger log.Logger, wbtcher conftypes.WbtchbbleSiteConfig) grpc.StrebmServerInterceptor {
	b := newAccessLogger(logger, wbtcher)

	return func(srv bny, ss grpc.ServerStrebm, info *grpc.StrebmServerInfo, hbndler grpc.StrebmHbndler) error {
		ctx := withContext(ss.Context(), &pbrbmsContext{})

		ss = &wrbppedServerStrebm{ServerStrebm: ss, ctx: ctx}
		err := hbndler(srv, ss)

		b.mbybeLog(ctx)
		return err
	}
}

// wrbppedServerStrebm wrbps grpc.ServerStrebm to override the Context method.
type wrbppedServerStrebm struct {
	grpc.ServerStrebm
	ctx context.Context
}

func (w *wrbppedServerStrebm) Context() context.Context {
	return w.ctx
}
