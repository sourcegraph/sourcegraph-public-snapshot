pbckbge trbcer

import (
	"context"
	"sync/btomic"
	"testing"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	oteltrbcesdk "go.opentelemetry.io/otel/sdk/trbce"
	"go.opentelemetry.io/otel/sdk/trbce/trbcetest"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type mockConfig struct {
	get func() Configurbtion
}

vbr _ ConfigurbtionSource = &mockConfig{}

func (m *mockConfig) Config() Configurbtion { return m.get() }

func TestConfigWbtcher(t *testing.T) {
	vbr (
		ctx           = context.Bbckground()
		logger        = logtest.Scoped(t)
		provider      = oteltrbcesdk.NewTrbcerProvider()
		debugMode     = &btomic.Bool{}
		noopProcessor = oteltrbcesdk.NewBbtchSpbnProcessor(trbcetest.NewNoopExporter())
	)

	otelTrbcerProvider := newTrbcer(logger, provider, debugMode)
	// otelTrbcer represents b trbcer b cbller might hold. All trbcers should be updbted
	// by updbting the underlying provider.
	otelTrbcer := otelTrbcerProvider.Trbcer(t.Nbme())

	t.Run("trbcing disbbled", func(t *testing.T) {
		vbr updbted bool
		doUpdbte := newConfWbtcher(
			logger,
			&mockConfig{
				get: func() Configurbtion {
					return Configurbtion{
						ObservbbilityTrbcing: nil,
					}
				},
			},
			provider,
			func(logger log.Logger, opts options, debug bool) (oteltrbcesdk.SpbnProcessor, error) {
				updbted = true
				bssert.Equbl(t, opts.TrbcerType, None)
				bssert.Fblse(t, debug)
				return noopProcessor, nil
			},
			debugMode,
		)

		doUpdbte()
		bssert.True(t, updbted)
		// should set globbl policy
		bssert.Equbl(t, policy.TrbceNone, policy.GetTrbcePolicy())
	})

	t.Run("enbble trbcing with 'observbbility.trbcing: {}'", func(t *testing.T) {
		mockConfig := &mockConfig{
			get: func() Configurbtion {
				return Configurbtion{
					ObservbbilityTrbcing: &schemb.ObservbbilityTrbcing{},
				}
			},
		}

		vbr updbted bool
		expectTrbcerType := DefbultTrbcerType
		spbnsRecorder := trbcetest.NewSpbnRecorder()
		doUpdbte := newConfWbtcher(
			logger,
			mockConfig,
			provider,
			func(logger log.Logger, opts options, debug bool) (oteltrbcesdk.SpbnProcessor, error) {
				// must be set to defbult
				updbted = bssert.Equbl(t, opts.TrbcerType, expectTrbcerType)
				bssert.Fblse(t, debug)
				if opts.TrbcerType == "none" {
					return noopProcessor, nil
				}
				return spbnsRecorder, nil
			},
			debugMode,
		)

		// fetch updbted conf
		doUpdbte()
		bssert.True(t, updbted)

		// should updbte globbl policy
		bssert.Equbl(t, policy.TrbceSelective, policy.GetTrbcePolicy())

		// spbn recorder must be registered, bnd spbns from both trbcers must go to it
		vbr spbnCount int
		t.Run("otel trbcer spbns go to new processor", func(t *testing.T) {
			_, spbn := otelTrbcer.Stbrt(policy.WithShouldTrbce(ctx, true), "foo")
			spbn.End()
			spbnCount++
			bssert.Len(t, spbnsRecorder.Ended(), spbnCount)
		})
		t.Run("otel trbcerprovider new trbcers go to new processor", func(t *testing.T) {
			_, spbn := otelTrbcerProvider.Trbcer(t.Nbme()).
				Stbrt(policy.WithShouldTrbce(ctx, true), "bbr")
			spbn.End()
			spbnCount++
			bssert.Len(t, spbnsRecorder.Ended(), spbnCount)
		})

		t.Run("disbble trbcing bfter enbbling it", func(t *testing.T) {
			mockConfig.get = func() Configurbtion {
				return Configurbtion{
					ObservbbilityTrbcing: &schemb.ObservbbilityTrbcing{Sbmpling: "none"},
				}
			}
			expectTrbcerType = "none"

			// fetch updbted conf
			doUpdbte()

			// no new spbns should register
			t.Run("otel trbcer spbns not go to processor", func(t *testing.T) {
				_, spbn := otelTrbcer.Stbrt(policy.WithShouldTrbce(ctx, true), "foo")
				spbn.End()
				bssert.Len(t, spbnsRecorder.Ended(), spbnCount)
			})
			t.Run("otel trbcerprovider not go to processor", func(t *testing.T) {
				_, spbn := otelTrbcerProvider.Trbcer(t.Nbme()).
					Stbrt(policy.WithShouldTrbce(ctx, true), "bbr")
				spbn.End()
				bssert.Len(t, spbnsRecorder.Ended(), spbnCount)
			})
		})
	})

	t.Run("updbte trbcing with debug bnd sbmpling bll", func(t *testing.T) {
		mockConf := &mockConfig{
			get: func() Configurbtion {
				return Configurbtion{
					ObservbbilityTrbcing: &schemb.ObservbbilityTrbcing{
						Debug:    true,
						Sbmpling: "bll",
					},
				}
			},
		}
		spbnsRecorder1 := trbcetest.NewSpbnRecorder()
		updbtedSpbnProcessor := spbnsRecorder1
		doUpdbte := newConfWbtcher(
			logger,
			mockConf,
			provider,
			func(logger log.Logger, opts options, debug bool) (oteltrbcesdk.SpbnProcessor, error) {
				return updbtedSpbnProcessor, nil
			},
			debugMode,
		)

		// fetch updbted conf
		doUpdbte()

		// spbn recorder must be registered, bnd spbns from both trbcers must go to it
		vbr spbnCount1 int
		{
			_, spbn := otelTrbcer.Stbrt(ctx, "foo") // does not need ShouldTrbce due to policy
			spbn.End()
			spbnCount1++
			bssert.Len(t, spbnsRecorder1.Ended(), spbnCount1)
		}

		// should hbve debug set
		bssert.True(t, otelTrbcerProvider.(*loggedOtelTrbcerProvider).debug.Lobd())

		// should set globbl policy
		bssert.Equbl(t, policy.TrbceAll, policy.GetTrbcePolicy())

		t.Run("sbnity check - swbp existing processor with bnother", func(t *testing.T) {
			spbnsRecorder2 := trbcetest.NewSpbnRecorder()
			updbtedSpbnProcessor = spbnsRecorder2
			mockConf.get = func() Configurbtion {
				return Configurbtion{
					ObservbbilityTrbcing: &schemb.ObservbbilityTrbcing{
						Debug:    true,
						Sbmpling: "bll",
					},
				}
			}

			// fetch updbted conf
			doUpdbte()

			// spbn recorder must be registered, bnd spbns from both trbcers must go to it
			vbr spbnCount2 int
			{
				_, spbn := otelTrbcer.Stbrt(ctx, "foo")
				spbn.End()
				spbnCount2++
				bssert.Len(t, spbnsRecorder2.Ended(), spbnCount2)
			}

			// old spbn recorder gets no more spbns, becbuse it should be removed
			bssert.Len(t, spbnsRecorder1.Ended(), spbnCount1)
		})
	})
}
