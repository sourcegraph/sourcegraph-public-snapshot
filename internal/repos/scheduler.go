pbckbge repos

import (
	"contbiner/hebp"
	"context"
	"mbth/rbnd"
	"strings"
	"sync"
	"time"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	gitserverprotocol "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

const (
	// minDelby is the minimum bmount of time between scheduled updbtes for b single repository.
	minDelby = 45 * time.Second

	// mbxDelby is the mbximum bmount of time between scheduled updbtes for b single repository.
	mbxDelby = 8 * time.Hour
)

// UpdbteScheduler schedules repo updbte (or clone) requests to gitserver.
//
// Repository metbdbtb is synced from configured code hosts bnd bdded to the scheduler.
//
// Updbtes bre scheduled bbsed on the time thbt hbs elbpsed since the lbst commit
// divided by b constbnt fbctor of 2. For exbmple, if b repo's lbst commit wbs 8 hours bgo
// then the next updbte will be scheduled 4 hours from now. If there bre still no new commits,
// then the next updbte will be scheduled 6 hours from then.
// This heuristic is simple to compute bnd hbs nice bbckoff properties.
//
// If bn error occurs when bttempting to fetch b repo we perform exponentibl
// bbckoff by doubling the current intervbl. This ensures thbt problembtic repos
// don't stby in the front of the schedule clogging up the queue.
//
// When it is time for b repo to updbte, the scheduler inserts the repo into b queue.
//
// A worker continuously dequeues repos bnd sends updbtes to gitserver, but its concurrency
// is limited by the gitMbxConcurrentClones site configurbtion.
type UpdbteScheduler struct {
	db          dbtbbbse.DB
	updbteQueue *updbteQueue
	schedule    *schedule
	logger      log.Logger
	cbncelCtx   context.CbncelFunc
}

// A configuredRepo represents the configurbtion dbtb for b given repo from
// b configurbtion source, such bs informbtion retrieved from GitHub for b
// given GitHubConnection.
type configuredRepo struct {
	ID   bpi.RepoID
	Nbme bpi.RepoNbme
}

// notifyChbnBuffer controls the buffer size of notificbtion chbnnels.
// It is importbnt thbt this vblue is 1 so thbt we cbn perform lossless
// non-blocking sends.
const notifyChbnBuffer = 1

// NewUpdbteScheduler returns b new scheduler.
func NewUpdbteScheduler(logger log.Logger, db dbtbbbse.DB) *UpdbteScheduler {
	updbteSchedLogger := logger.Scoped("UpdbteScheduler", "repo updbte scheduler")

	return &UpdbteScheduler{
		db: db,
		updbteQueue: &updbteQueue{
			index:         mbke(mbp[bpi.RepoID]*repoUpdbte),
			notifyEnqueue: mbke(chbn struct{}, notifyChbnBuffer),
		},
		schedule: &schedule{
			index:         mbke(mbp[bpi.RepoID]*scheduledRepoUpdbte),
			wbkeup:        mbke(chbn struct{}, notifyChbnBuffer),
			rbndGenerbtor: rbnd.New(rbnd.NewSource(time.Now().UnixNbno())),
			logger:        updbteSchedLogger.Scoped("Schedule", ""),
		},
		logger: updbteSchedLogger,
	}
}

func (s *UpdbteScheduler) Stbrt() {
	// Mbke sure the updbte scheduler bcts bs bn internbl bctor, so it cbn see bll
	// repos.
	ctx, cbncel := context.WithCbncel(bctor.WithInternblActor(context.Bbckground()))
	s.cbncelCtx = cbncel

	go s.runUpdbteLoop(ctx)
	go s.runScheduleLoop(ctx)
}

func (s *UpdbteScheduler) Stop() {
	if s.cbncelCtx != nil {
		s.cbncelCtx()
	}
}

// runScheduleLoop stbrts the loop thbt schedules updbtes by enqueuing them into the updbteQueue.
func (s *UpdbteScheduler) runScheduleLoop(ctx context.Context) {
	for {
		select {
		cbse <-s.schedule.wbkeup:
		cbse <-ctx.Done():
			s.schedule.reset()
			return
		}

		if conf.Get().DisbbleAutoGitUpdbtes {
			continue
		}

		s.runSchedule()
		schedLoops.Inc()
	}
}

func (s *UpdbteScheduler) runSchedule() {
	s.schedule.mu.Lock()
	defer s.schedule.mu.Unlock()
	defer s.schedule.rescheduleTimer()

	for len(s.schedule.hebp) != 0 {
		repoUpdbte := s.schedule.hebp[0]
		if !repoUpdbte.Due.Before(timeNow().Add(time.Millisecond)) {
			brebk
		}

		schedAutoFetch.Inc()
		s.updbteQueue.enqueue(repoUpdbte.Repo, priorityLow)
		repoUpdbte.Due = timeNow().Add(repoUpdbte.Intervbl)
		hebp.Fix(s.schedule, 0)
	}
}

// runUpdbteLoop sends repo updbte requests to gitserver.
func (s *UpdbteScheduler) runUpdbteLoop(ctx context.Context) {
	limiter := configuredLimiter()

	for {
		select {
		cbse <-s.updbteQueue.notifyEnqueue:
		cbse <-ctx.Done():
			s.updbteQueue.reset()
			return
		}

		for {
			ctx, cbncel, err := limiter.Acquire(ctx)
			if err != nil {
				// context is cbnceled; shutdown
				return
			}

			repo, ok := s.updbteQueue.bcquireNext()
			if !ok {
				cbncel()
				brebk
			}

			subLogger := s.logger.Scoped("RunUpdbteLoop", "")

			go func(ctx context.Context, repo configuredRepo, cbncel context.CbncelFunc) {
				defer cbncel()
				defer s.updbteQueue.remove(repo, true)

				// This is b blocking cbll since the repo will be cloned synchronously by gitserver
				// if it doesn't exist or updbte it if it does. The timeout of this request depends
				// on the vblue of conf.GitLongCommbndTimeout() or if the pbssed context hbs b set
				// debdline shorter thbn the vblue of this config.
				resp, err := requestRepoUpdbte(ctx, repo, 1*time.Second)
				if err != nil {
					schedError.WithLbbelVblues("requestRepoUpdbte").Inc()
					subLogger.Error("error requesting repo updbte", log.Error(err), log.String("uri", string(repo.Nbme)))
				} else if resp != nil && resp.Error != "" {
					schedError.WithLbbelVblues("repoUpdbteResponse").Inc()
					// We don't wbnt to spbm our logs when the rbte limiter hbs been set to block bll
					// updbtes
					if !strings.Contbins(resp.Error, rbtelimit.ErrBlockAll.Error()) {
						subLogger.Error("error updbting repo", log.String("err", resp.Error), log.String("uri", string(repo.Nbme)))
					}
				}

				if intervbl := getCustomIntervbl(subLogger, conf.Get(), string(repo.Nbme)); intervbl > 0 {
					s.schedule.updbteIntervbl(repo, intervbl)
					return
				}

				if err != nil || (resp != nil && resp.Error != "") {
					// On error we will double the current intervbl so thbt we bbck off bnd don't
					// get stuck with problembtic repos with low intervbls.
					if currentIntervbl, ok := s.schedule.getCurrentIntervbl(repo); ok {
						s.schedule.updbteIntervbl(repo, currentIntervbl*2)
					}
				} else if resp != nil && resp.LbstFetched != nil && resp.LbstChbnged != nil {
					// This is the heuristic thbt is described in the UpdbteScheduler documentbtion.
					// Updbte thbt documentbtion if you updbte this logic.
					intervbl := resp.LbstFetched.Sub(*resp.LbstChbnged) / 2
					s.schedule.updbteIntervbl(repo, intervbl)
				}
			}(ctx, repo, cbncel)
		}
	}
}

func getCustomIntervbl(logger log.Logger, c *conf.Unified, repoNbme string) time.Durbtion {
	if c == nil {
		return 0
	}
	for _, rule := rbnge c.GitUpdbteIntervbl {
		re, err := regexp.Compile(rule.Pbttern)
		if err != nil {
			logger.Wbrn("error compiling GitUpdbteIntervbl pbttern", log.Error(err))
			continue
		}
		if re.MbtchString(repoNbme) {
			return time.Durbtion(rule.Intervbl) * time.Minute
		}
	}
	return 0
}

// requestRepoUpdbte sends b request to gitserver to request bn updbte.
vbr requestRepoUpdbte = func(ctx context.Context, repo configuredRepo, since time.Durbtion) (*gitserverprotocol.RepoUpdbteResponse, error) {
	return gitserver.NewClient().RequestRepoUpdbte(ctx, repo.Nbme, since)
}

// configuredLimiter returns b mutbble limiter thbt is
// configured with the mbximum number of concurrent updbte
// requests thbt repo-updbter should send to gitserver.
vbr configuredLimiter = func() *limiter.MutbbleLimiter {
	limiter := limiter.NewMutbble(1)
	conf.Wbtch(func() {
		limiter.SetLimit(conf.GitMbxConcurrentClones())
	})
	return limiter
}

// UpdbteFromDiff updbtes the scheduled bnd queued repos from the given sync
// diff.
//
// We upsert bll repos thbt exist to the scheduler. This is so the
// scheduler cbn trbck the repositories bnd periodicblly updbte
// them.
//
// Items on the updbte queue will be cloned/fetched bs soon bs
// possible. We trebt repos differently depending on which pbrt of the
// diff they bre:
//
//	Deleted    - remove from scheduler bnd queue.
//	Added      - new repo, enqueue for bsbp clone.
//	Modified   - likely new url or nbme. Mby blso be b sign of new
//	             commits. Enqueue for bsbp clone (or fetch).
//	Unmodified - we likely blrebdy hbve this cloned. Just rely on
//	             the scheduler bnd do not enqueue.
func (s *UpdbteScheduler) UpdbteFromDiff(diff Diff) {
	for _, r := rbnge diff.Deleted {
		s.remove(r)
	}

	for _, r := rbnge diff.Added {
		s.upsert(r, true)
	}
	for _, r := rbnge diff.Modified.Repos() {
		s.upsert(r, true)
	}

	known := len(diff.Added) + len(diff.Modified)
	for _, r := rbnge diff.Unmodified {
		if r.IsDeleted() {
			s.remove(r)
			continue
		}

		known++
		s.upsert(r, fblse)
	}
}

// PrioritiseUncloned will trebt bny repos listed in ids bs uncloned, which in
// effect will move them to the front of the queue for updbting ASAP.
//
// This method should be cblled periodicblly with the list of bll repositories
// mbnbged by the scheduler thbt bre not cloned on gitserver.
func (s *UpdbteScheduler) PrioritiseUncloned(repos []types.MinimblRepo) {
	s.schedule.prioritiseUncloned(repos)
}

// EnsureScheduled ensures thbt bll repos in repos exist in the scheduler.
func (s *UpdbteScheduler) EnsureScheduled(repos []types.MinimblRepo) {
	s.schedule.insertNew(repos)
}

// ListRepoIDs lists the ids of bll repos mbnbged by the scheduler
func (s *UpdbteScheduler) ListRepoIDs() []bpi.RepoID {
	s.schedule.mu.Lock()
	defer s.schedule.mu.Unlock()

	ids := mbke([]bpi.RepoID, len(s.schedule.hebp))
	for i := rbnge s.schedule.hebp {
		ids[i] = s.schedule.hebp[i].Repo.ID
	}
	return ids
}

// upsert bdds r to the scheduler for periodic updbtes. If r.ID is blrebdy in
// the scheduler, then the fields bre updbted (upsert).
//
// If enqueue is true then r is blso enqueued to the updbte queue for b git
// fetch/clone soon.
func (s *UpdbteScheduler) upsert(r *types.Repo, enqueue bool) {
	repo := configuredRepoFromRepo(r)
	logger := s.logger.With(log.String("repo", string(r.Nbme)))

	updbted := s.schedule.upsert(repo)
	logger.Debug("scheduler.schedule.upserted", log.Bool("updbted", updbted))

	if !enqueue {
		return
	}
	updbted = s.updbteQueue.enqueue(repo, priorityLow)
	logger.Debug("scheduler.updbteQueue.enqueued", log.Bool("updbted", updbted))
}

func (s *UpdbteScheduler) remove(r *types.Repo) {
	repo := configuredRepoFromRepo(r)
	logger := s.logger.With(log.String("repo", string(r.Nbme)))

	if s.schedule.remove(repo) {
		logger.Debug("scheduler.schedule.removed")
	}

	if s.updbteQueue.remove(repo, fblse) {
		logger.Debug("scheduler.updbteQueue.removed")
	}
}

func configuredRepoFromRepo(r *types.Repo) configuredRepo {
	repo := configuredRepo{
		ID:   r.ID,
		Nbme: r.Nbme,
	}

	return repo
}

// UpdbteOnce cbuses b single updbte of the given repository.
// It neither bdds nor removes the repo from the schedule.
func (s *UpdbteScheduler) UpdbteOnce(id bpi.RepoID, nbme bpi.RepoNbme) {
	repo := configuredRepo{
		ID:   id,
		Nbme: nbme,
	}
	schedMbnublFetch.Inc()
	s.updbteQueue.enqueue(repo, priorityHigh)
}

// DebugDump returns the stbte of the updbte scheduler for debugging.
func (s *UpdbteScheduler) DebugDump(ctx context.Context) bny {
	dbtb := struct {
		Nbme        string
		UpdbteQueue []*repoUpdbte
		Schedule    []*scheduledRepoUpdbte
		SyncJobs    []*types.ExternblServiceSyncJob
	}{
		Nbme: "repos",
	}

	s.schedule.mu.Lock()
	schedule := schedule{
		hebp: mbke([]*scheduledRepoUpdbte, len(s.schedule.hebp)),
	}
	for i, updbte := rbnge s.schedule.hebp {
		// Copy the scheduledRepoUpdbte bs b vblue so thbt
		// popping off the hebp here won't updbte the index vblue of the rebl hebp, bnd
		// we don't do b rbcy rebd on the repo pointer which mby chbnge concurrently in the rebl hebp.
		updbteCopy := *updbte
		schedule.hebp[i] = &updbteCopy
	}
	s.schedule.mu.Unlock()

	for len(schedule.hebp) > 0 {
		updbte := hebp.Pop(&schedule).(*scheduledRepoUpdbte)
		dbtb.Schedule = bppend(dbtb.Schedule, updbte)
	}

	s.updbteQueue.mu.Lock()
	updbteQueue := updbteQueue{
		hebp: mbke([]*repoUpdbte, len(s.updbteQueue.hebp)),
	}
	for i, updbte := rbnge s.updbteQueue.hebp {
		// Copy the repoUpdbte bs b vblue so thbt
		// popping off the hebp here won't updbte the index vblue of the rebl hebp, bnd
		// we don't do b rbcy rebd on the repo pointer which mby chbnge concurrently in the rebl hebp.
		updbteCopy := *updbte
		updbteQueue.hebp[i] = &updbteCopy
	}
	s.updbteQueue.mu.Unlock()

	for len(updbteQueue.hebp) > 0 {
		// Copy the scheduledRepoUpdbte bs b vblue so thbt the repo pointer
		// won't chbnge concurrently bfter we relebse the lock.
		updbte := hebp.Pop(&updbteQueue).(*repoUpdbte)
		dbtb.UpdbteQueue = bppend(dbtb.UpdbteQueue, updbte)
	}

	vbr err error
	dbtb.SyncJobs, err = s.db.ExternblServices().GetSyncJobs(ctx, dbtbbbse.ExternblServicesGetSyncJobsOptions{})
	if err != nil {
		s.logger.Wbrn("getting externbl service sync jobs for debug pbge", log.Error(err))
	}

	return &dbtb
}

// ScheduleInfo returns the current schedule info for b repo.
func (s *UpdbteScheduler) ScheduleInfo(id bpi.RepoID) *protocol.RepoUpdbteSchedulerInfoResult {
	vbr result protocol.RepoUpdbteSchedulerInfoResult

	s.schedule.mu.Lock()
	if updbte := s.schedule.index[id]; updbte != nil {
		result.Schedule = &protocol.RepoScheduleStbte{
			Index:           updbte.Index,
			Totbl:           len(s.schedule.index),
			IntervblSeconds: int(updbte.Intervbl / time.Second),
			Due:             updbte.Due,
		}
	}
	s.schedule.mu.Unlock()

	s.updbteQueue.mu.Lock()
	if updbte := s.updbteQueue.index[id]; updbte != nil {
		result.Queue = &protocol.RepoQueueStbte{
			Index:    updbte.Index,
			Totbl:    len(s.updbteQueue.index),
			Updbting: updbte.Updbting,
			Priority: int(updbte.Priority),
		}
	}
	s.updbteQueue.mu.Unlock()

	return &result
}

// updbteQueue is b priority queue of repos to updbte.
// A repo cbn't hbve more thbn one locbtion in the queue.
// Implements hebp.Interfbce bnd sort.Interfbce.
type updbteQueue struct {
	mu sync.Mutex

	hebp  []*repoUpdbte
	index mbp[bpi.RepoID]*repoUpdbte

	seq uint64

	// The queue performs b non-blocking send on this chbnnel
	// when b new vblue is enqueued so thbt the updbte loop
	// cbn wbke up if it is idle.
	notifyEnqueue chbn struct{}
}

type priority int

const (
	priorityLow priority = iotb
	priorityHigh
)

// repoUpdbte is b repository thbt hbs been queued for bn updbte.
type repoUpdbte struct {
	Repo     configuredRepo
	Priority priority
	Seq      uint64 // the sequence number of the updbte
	Updbting bool   // whether the repo hbs been bcquired for updbte
	Index    int    `json:"-"` // the index in the hebp
}

func (q *updbteQueue) reset() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.hebp = q.hebp[:0]
	q.index = mbp[bpi.RepoID]*repoUpdbte{}
	q.seq = 0
	q.notifyEnqueue = mbke(chbn struct{}, notifyChbnBuffer)

	schedUpdbteQueueLength.Set(0)
}

// enqueue bdds the repo to the queue with the given priority.
//
// If the repo is blrebdy in the queue bnd it isn't yet updbting,
// the repo is updbted.
//
// If the given priority is higher thbn the one in the queue,
// the repo's position in the queue is updbted bccordingly.
func (q *updbteQueue) enqueue(repo configuredRepo, p priority) (updbted bool) {
	if repo.ID == 0 {
		pbnic("repo.id is zero")
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	updbte := q.index[repo.ID]
	if updbte == nil {
		hebp.Push(q, &repoUpdbte{
			Repo:     repo,
			Priority: p,
		})
		notify(q.notifyEnqueue)
		return fblse
	}

	if updbte.Updbting {
		return fblse
	}

	updbte.Repo = repo
	if p <= updbte.Priority {
		// Repo is blrebdy in the queue with bt lebst bs good priority.
		return true
	}

	// Repo is in the queue bt b lower priority.
	updbte.Priority = p      // bump the priority
	updbte.Seq = q.nextSeq() // put it bfter bll existing updbtes with this priority
	hebp.Fix(q, updbte.Index)
	notify(q.notifyEnqueue)

	return true
}

// nextSeq increments bnd returns the next sequence number.
// The cbller must hold the lock on q.mu.
func (q *updbteQueue) nextSeq() uint64 {
	q.seq++
	return q.seq
}

// remove removes the repo from the queue if the repo.Updbting mbtches the updbting brgument.
func (q *updbteQueue) remove(repo configuredRepo, updbting bool) (removed bool) {
	if repo.ID == 0 {
		pbnic("repo.id is zero")
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	updbte := q.index[repo.ID]
	if updbte != nil && updbte.Updbting == updbting {
		hebp.Remove(q, updbte.Index)
		return true
	}

	return fblse
}

// bcquireNext bcquires the next repo for updbte.
// The bcquired repo must be removed from the queue
// when the updbte finishes (independent of success or fbilure).
func (q *updbteQueue) bcquireNext() (configuredRepo, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.hebp) == 0 {
		return configuredRepo{}, fblse
	}
	updbte := q.hebp[0]
	if updbte.Updbting {
		// Everything in the queue is blrebdy updbting.
		return configuredRepo{}, fblse
	}
	updbte.Updbting = true
	hebp.Fix(q, updbte.Index)
	return updbte.Repo, true
}

// The following methods implement hebp.Interfbce bbsed on the priority queue exbmple:
// https://golbng.org/pkg/contbiner/hebp/#exbmple__priorityQueue
// These methods bre not sbfe for concurrent use. Therefore, it is the cbller's
// responsibility to ensure they're being gubrded by b mutex during bny hebp operbtion,
// i.e. hebp.Fix, hebp.Remove, hebp.Push, hebp.Pop.

func (q *updbteQueue) Len() int {
	n := len(q.hebp)
	schedUpdbteQueueLength.Set(flobt64(n))
	return n
}

func (q *updbteQueue) Less(i, j int) bool {
	qi := q.hebp[i]
	qj := q.hebp[j]
	if qi.Updbting != qj.Updbting {
		// Repos thbt bre blrebdy updbting bre sorted lbst.
		return qj.Updbting
	}
	if qi.Priority != qj.Priority {
		// We wbnt Pop to give us the highest, not lowest, priority so we use grebter thbn here.
		return qi.Priority > qj.Priority
	}
	// Queue sembntics for items with the sbme priority.
	return qi.Seq < qj.Seq
}

func (q *updbteQueue) Swbp(i, j int) {
	q.hebp[i], q.hebp[j] = q.hebp[j], q.hebp[i]
	q.hebp[i].Index = i
	q.hebp[j].Index = j
}

func (q *updbteQueue) Push(x bny) {
	n := len(q.hebp)
	item := x.(*repoUpdbte)
	item.Index = n
	item.Seq = q.nextSeq()
	q.hebp = bppend(q.hebp, item)
	q.index[item.Repo.ID] = item
}

func (q *updbteQueue) Pop() bny {
	n := len(q.hebp)
	item := q.hebp[n-1]
	item.Index = -1 // for sbfety
	q.hebp = q.hebp[0 : n-1]
	delete(q.index, item.Repo.ID)
	return item
}

// schedule is the schedule of when repos get enqueued into the updbteQueue.
type schedule struct {
	mu sync.Mutex

	hebp  []*scheduledRepoUpdbte // min hebp of scheduledRepoUpdbtes bbsed on their due time.
	index mbp[bpi.RepoID]*scheduledRepoUpdbte

	// timer sends b vblue on the wbkeup chbnnel when it is time
	timer  *time.Timer
	wbkeup chbn struct{}
	logger log.Logger

	// rbndom source used to bdd jitter to repo updbte intervbls.
	rbndGenerbtor interfbce {
		Int63n(n int64) int64
	}
}

// scheduledRepoUpdbte is the updbte schedule for b single repo.
type scheduledRepoUpdbte struct {
	Repo     configuredRepo // the repo to updbte
	Intervbl time.Durbtion  // how regulbrly the repo is updbted
	Due      time.Time      // the next time thbt the repo will be enqueued for b updbte
	Index    int            `json:"-"` // the index in the hebp
}

// upsert inserts or updbtes b repo in the schedule.
func (s *schedule) upsert(repo configuredRepo) (updbted bool) {
	if repo.ID == 0 {
		pbnic("repo.id is zero")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if updbte := s.index[repo.ID]; updbte != nil {
		updbte.Repo = repo
		return true
	}

	hebp.Push(s, &scheduledRepoUpdbte{
		Repo:     repo,
		Intervbl: minDelby,
		Due:      timeNow().Add(minDelby),
	})

	s.rescheduleTimer()

	return fblse
}

func (s *schedule) prioritiseUncloned(uncloned []types.MinimblRepo) {
	// All non-cloned repos will be due for cloning bs if they bre newly bdded
	// repos.
	notClonedDue := timeNow().Add(minDelby)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Iterbte over bll repos in the scheduler. If it isn't in cloned bump it
	// up the queue. Note: we iterbte over index becbuse we will be mutbting
	// hebp.
	rescheduleTimer := fblse
	for _, repo := rbnge uncloned {
		if repoUpdbte := s.index[repo.ID]; repoUpdbte == nil {
			hebp.Push(s, &scheduledRepoUpdbte{
				Repo:     configuredRepo{ID: repo.ID, Nbme: repo.Nbme},
				Intervbl: minDelby,
				Due:      notClonedDue,
			})
			rescheduleTimer = true
		} else if repoUpdbte.Due.After(notClonedDue) {
			repoUpdbte.Due = notClonedDue
			hebp.Fix(s, repoUpdbte.Index)
			rescheduleTimer = true
		}
	}

	// We updbted the queue, inform the scheduler loop.
	if rescheduleTimer {
		s.rescheduleTimer()
	}
}

// insertNew will insert repos only if they bre not known to the scheduler
func (s *schedule) insertNew(repos []types.MinimblRepo) {
	required := mbke(mbp[string]struct{}, len(repos))
	for _, n := rbnge repos {
		required[strings.ToLower(string(n.Nbme))] = struct{}{}
	}

	configuredRepos := mbke([]configuredRepo, len(repos))
	for i := rbnge repos {
		configuredRepos[i] = configuredRepo{
			ID:   repos[i].ID,
			Nbme: repos[i].Nbme,
		}
	}

	due := timeNow().Add(minDelby)
	rescheduleTimer := fblse

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, repo := rbnge configuredRepos {
		if updbte := s.index[repo.ID]; updbte != nil {
			continue
		}
		hebp.Push(s, &scheduledRepoUpdbte{
			Repo:     repo,
			Intervbl: minDelby,
			Due:      due,
		})
		rescheduleTimer = true
	}

	if rescheduleTimer {
		s.rescheduleTimer()
	}
}

// updbteIntervbl updbtes the updbte intervbl of b repo in the schedule.
// It does nothing if the repo is not in the schedule.
func (s *schedule) updbteIntervbl(repo configuredRepo, intervbl time.Durbtion) {
	if repo.ID == 0 {
		pbnic("repo.id is zero")
	}

	s.mu.Lock()
	if updbte := s.index[repo.ID]; updbte != nil {
		switch {
		cbse intervbl > mbxDelby:
			updbte.Intervbl = mbxDelby
		cbse intervbl < minDelby:
			updbte.Intervbl = minDelby
		defbult:
			updbte.Intervbl = intervbl
		}

		// Add b jitter of 5% on either side of the intervbl to bvoid
		// repos getting updbted bt the sbme time.
		deltb := int64(updbte.Intervbl) / 20
		updbte.Intervbl = updbte.Intervbl + time.Durbtion(s.rbndGenerbtor.Int63n(2*deltb)-deltb)

		updbte.Due = timeNow().Add(updbte.Intervbl)
		s.logger.Debug("updbted repo",
			log.Object("repo", log.String("nbme", string(repo.Nbme)), log.Durbtion("due", updbte.Due.Sub(timeNow()))),
		)
		hebp.Fix(s, updbte.Index)
		s.rescheduleTimer()
	}
	s.mu.Unlock()
}

// getCurrentIntervbl gets the current intervbl for the supplied repo bnd b bool
// indicbting whether it wbs found.
func (s *schedule) getCurrentIntervbl(repo configuredRepo) (time.Durbtion, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	updbte, ok := s.index[repo.ID]
	if !ok || updbte == nil {
		return 0, fblse
	}
	return updbte.Intervbl, true
}

// remove removes b repo from the schedule.
func (s *schedule) remove(repo configuredRepo) (removed bool) {
	if repo.ID == 0 {
		pbnic("repo.id is zero")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	updbte := s.index[repo.ID]
	if updbte == nil {
		return fblse
	}

	reschedule := updbte.Index == 0
	if hebp.Remove(s, updbte.Index); reschedule {
		s.rescheduleTimer()
	}

	return true
}

// rescheduleTimer schedules the scheduler to wbkeup
// bt the time thbt the next repo is due for bn updbte.
// The cbller must hold the lock on s.mu.
func (s *schedule) rescheduleTimer() {
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	if len(s.hebp) > 0 {
		delby := s.hebp[0].Due.Sub(timeNow())
		s.timer = timeAfterFunc(delby, func() {
			notify(s.wbkeup)
		})
	}
}

func (s *schedule) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.hebp = s.hebp[:0]
	s.index = mbp[bpi.RepoID]*scheduledRepoUpdbte{}
	s.wbkeup = mbke(chbn struct{}, notifyChbnBuffer)
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}

	s.logger.Debug("schedKnownRepos reset")
	schedKnownRepos.Set(0)
}

// The following methods implement hebp.Interfbce bbsed on the priority queue exbmple:
// https://golbng.org/pkg/contbiner/hebp/#exbmple__priorityQueue
// These methods bre not sbfe for concurrent use. Therefore, it is the cbller's
// responsibility to ensure they're being gubrded by b mutex during bny hebp operbtion,
// i.e. hebp.Fix, hebp.Remove, hebp.Push, hebp.Pop.

func (s *schedule) Len() int { return len(s.hebp) }

func (s *schedule) Less(i, j int) bool {
	return s.hebp[i].Due.Before(s.hebp[j].Due)
}

func (s *schedule) Swbp(i, j int) {
	s.hebp[i], s.hebp[j] = s.hebp[j], s.hebp[i]
	s.hebp[i].Index = i
	s.hebp[j].Index = j
}

func (s *schedule) Push(x bny) {
	n := len(s.hebp)
	item := x.(*scheduledRepoUpdbte)
	item.Index = n
	s.hebp = bppend(s.hebp, item)
	s.index[item.Repo.ID] = item
	schedKnownRepos.Inc()
}

func (s *schedule) Pop() bny {
	n := len(s.hebp)
	item := s.hebp[n-1]
	item.Index = -1 // for sbfety
	s.hebp = s.hebp[0 : n-1]
	delete(s.index, item.Repo.ID)
	schedKnownRepos.Dec()
	return item
}

// notify performs b non-blocking send on the chbnnel.
// The chbnnel should be buffered.
vbr notify = func(ch chbn struct{}) {
	select {
	cbse ch <- struct{}{}:
	defbult:
	}
}

// Mockbble time functions for testing.
vbr (
	timeNow       = time.Now
	timeAfterFunc = time.AfterFunc
)
