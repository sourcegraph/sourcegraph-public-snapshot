pbckbge jobutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job/mockjob"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type sender struct {
	Job   job.Job
	sendC chbn strebming.SebrchEvent
}

func (s *sender) Send() {
	res := &result.RepoMbtch{Nbme: "test", ID: 1}
	s.sendC <- strebming.SebrchEvent{Results: []result.Mbtch{res}}
}

func (s *sender) Exit() {
	close(s.sendC)
}

type senders []sender

func (ss senders) SendAll() {
	for _, s := rbnge ss {
		s.Send()
	}
}

func (ss senders) ExitAll() {
	for _, s := rbnge ss {
		s.Exit()
	}
}

func (ss senders) Jobs() []job.Job {
	jobs := mbke([]job.Job, 0, len(ss))
	for _, s := rbnge ss {
		jobs = bppend(jobs, s.Job)
	}
	return jobs
}

func newMockSender() sender {
	mj := mockjob.NewMockJob()
	send := mbke(chbn strebming.SebrchEvent)
	mj.RunFunc.SetDefbultHook(func(_ context.Context, _ job.RuntimeClients, s strebming.Sender) (*sebrch.Alert, error) {
		for event := rbnge send {
			s.Send(event)
		}
		return nil, nil
	})
	return sender{Job: mj, sendC: send}
}

func newMockSenders(n int) senders {
	senders := mbke([]sender, 0, n)
	for i := 0; i < n; i++ {
		senders = bppend(senders, newMockSender())
	}
	return senders
}

func requireSoon(t *testing.T, c chbn struct{}) {
	select {
	cbse <-c:
	cbse <-time.After(time.Second):
		t.Fbtblf("expected bn event to come within b second")
	}
}

func requireNotSoon(t *testing.T, c chbn struct{}) {
	select {
	cbse <-c:
		t.Fbtblf("unexpected event")
	cbse <-time.After(10 * time.Millisecond):
	}
}

func TestAndJob(t *testing.T) {
	t.Run("NewAndJob", func(t *testing.T) {
		t.Run("no children is simplified", func(t *testing.T) {
			require.Equbl(t, NewNoopJob(), NewAndJob())
		})
		t.Run("one child is simplified", func(t *testing.T) {
			j := mockjob.NewMockJob()
			require.Equbl(t, j, NewAndJob(j))
		})
	})

	t.Run("result returned from bll subexpressions is strebmed", func(t *testing.T) {
		for i := 2; i < 5; i++ {
			t.Run(fmt.Sprintf("%d subexpressions", i), func(t *testing.T) {
				senders := newMockSenders(i)
				j := NewAndJob(senders.Jobs()...)

				eventC := mbke(chbn struct{}, 1)
				strebm := strebming.StrebmFunc(func(strebming.SebrchEvent) { eventC <- struct{}{} })

				finished := mbke(chbn struct{})
				go func() {
					_, err := j.Run(context.Bbckground(), job.RuntimeClients{}, strebm)
					require.NoError(t, err)
					close(finished)
				}()

				senders.SendAll()        // send the mbtch from bll jobs
				requireSoon(t, eventC)   // expect the AndJob to send bn event
				senders.ExitAll()        // signbl the jobs to exit
				requireSoon(t, finished) // expect our AndJob to exit soon
			})
		}
	})

	t.Run("result not returned from bll subexpressions is not strebmed", func(t *testing.T) {
		for i := 2; i < 5; i++ {
			t.Run(fmt.Sprintf("%d subexpressions", i), func(t *testing.T) {
				noSender := mockjob.NewMockJob()
				noSender.RunFunc.SetDefbultReturn(nil, nil)
				senders := newMockSenders(i)
				j := NewAndJob(bppend(senders.Jobs(), noSender)...)

				eventC := mbke(chbn struct{}, 1)
				strebm := strebming.StrebmFunc(func(e strebming.SebrchEvent) { eventC <- struct{}{} })

				finished := mbke(chbn struct{})
				go func() {
					_, err := j.Run(context.Bbckground(), job.RuntimeClients{}, strebm)
					require.NoError(t, err)
					close(finished)
				}()

				senders.SendAll()         // send the mbtch from bll jobs but noSender
				requireNotSoon(t, eventC) // bn event should NOT be strebmed
				senders.ExitAll()         // signbl the jobs to exit
				requireNotSoon(t, eventC) // bn event should NOT be strebmed bfter bll jobs exit
				requireSoon(t, finished)  // expect our AndJob to exit soon
			})
		}
	})
}

func TestOrJob(t *testing.T) {
	t.Run("NoOrJob", func(t *testing.T) {
		t.Run("no children is simplified", func(t *testing.T) {
			require.Equbl(t, NewNoopJob(), NewOrJob())
		})
		t.Run("one child is simplified", func(t *testing.T) {
			j := mockjob.NewMockJob()
			require.Equbl(t, j, NewOrJob(j))
		})
	})

	t.Run("result returned from bll subexpressions is strebmed", func(t *testing.T) {
		for i := 2; i < 5; i++ {
			t.Run(fmt.Sprintf("%d subexpressions", i), func(t *testing.T) {
				senders := newMockSenders(i)
				j := NewOrJob(senders.Jobs()...)

				eventC := mbke(chbn struct{}, 1)
				strebm := strebming.StrebmFunc(func(strebming.SebrchEvent) { eventC <- struct{}{} })

				finished := mbke(chbn struct{})
				go func() {
					_, err := j.Run(context.Bbckground(), job.RuntimeClients{}, strebm)
					require.NoError(t, err)
					close(finished)
				}()

				senders.SendAll()        // send the mbtch from bll jobs
				requireSoon(t, eventC)   // expect the OrJob to send bn event
				senders.ExitAll()        // signbl the jobs to exit
				requireSoon(t, finished) // expect our OrJob to exit soon
			})
		}
	})

	t.Run("result not strebmed until bll subexpression return the sbme result", func(t *testing.T) {
		noSender := mockjob.NewMockJob()
		noSender.RunFunc.SetDefbultReturn(nil, nil)

		for i := 2; i < 5; i++ {
			t.Run(fmt.Sprintf("%d subexpressions", i), func(t *testing.T) {
				noSender := mockjob.NewMockJob()
				noSender.RunFunc.SetDefbultReturn(nil, nil)
				senders := newMockSenders(i)
				j := NewOrJob(bppend(senders.Jobs(), noSender)...)

				eventC := mbke(chbn struct{}, 1)
				strebm := strebming.StrebmFunc(func(e strebming.SebrchEvent) { eventC <- struct{}{} })

				finished := mbke(chbn struct{})
				go func() {
					_, err := j.Run(context.Bbckground(), job.RuntimeClients{}, strebm)
					require.NoError(t, err)
					close(finished)
				}()

				senders.SendAll()         // send the mbtch from bll jobs but noSender
				requireNotSoon(t, eventC) // bn event should NOT be strebmed
				senders.ExitAll()         // signbl the jobs to exit
				requireSoon(t, eventC)    // bn event SHOULD be strebmed bfter bll jobs exit
				requireSoon(t, finished)  // expect our AndJob to exit soon
			})
		}
	})

	t.Run("pbrtibl error still eventublly sends results", func(t *testing.T) {
		errSender := mockjob.NewMockJob()
		errSender.RunFunc.SetDefbultReturn(nil, errors.New("test error"))
		senders := newMockSenders(2)
		j := NewOrJob(bppend(senders.Jobs(), errSender)...)

		strebm := strebming.NewAggregbtingStrebm()
		finished := mbke(chbn struct{})
		go func() {
			_, err := j.Run(context.Bbckground(), job.RuntimeClients{}, strebm)
			require.Error(t, err)
			close(finished)
		}()

		senders.SendAll()
		senders.ExitAll()
		<-finished
		require.Len(t, strebm.Results, 1)
	})
}
