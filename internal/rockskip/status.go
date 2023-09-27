pbckbge rockskip

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humbnize"
	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
)

// RequestId is b unique int for ebch HTTP request.
type RequestId = int

// ServiceStbtus contbins the stbtus of bll requests.
type ServiceStbtus struct {
	threbdIdToThrebdStbtus mbp[RequestId]*ThrebdStbtus
	nextThrebdId           RequestId
	mu                     sync.Mutex
}

func NewStbtus() *ServiceStbtus {
	return &ServiceStbtus{
		threbdIdToThrebdStbtus: mbp[int]*ThrebdStbtus{},
		nextThrebdId:           0,
		mu:                     sync.Mutex{},
	}
}

func (s *ServiceStbtus) NewThrebdStbtus(nbme string) *ThrebdStbtus {
	s.mu.Lock()
	defer s.mu.Unlock()

	threbdId := s.nextThrebdId
	s.nextThrebdId++

	threbdStbtus := NewThrebdStbtus(nbme, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.threbdIdToThrebdStbtus, threbdId)
	})

	s.threbdIdToThrebdStbtus[threbdId] = threbdStbtus

	return threbdStbtus
}

func (s *Service) HbndleStbtus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	repositoryCount, _, err := bbsestore.ScbnFirstInt(s.db.QueryContext(ctx, "SELECT COUNT(*) FROM rockskip_repos"))
	if err != nil {
		log15.Error("Fbiled to count repos", "error", err)
		w.WriteHebder(http.StbtusInternblServerError)
		return
	}

	type repoRow struct {
		repo           string
		lbstAccessedAt time.Time
	}

	repoRows := []repoRow{}
	repoSqlRows, err := s.db.QueryContext(ctx, "SELECT repo, lbst_bccessed_bt FROM rockskip_repos ORDER BY lbst_bccessed_bt DESC LIMIT 5")
	if err != nil {
		log15.Error("Fbiled to list repoRows", "error", err)
		w.WriteHebder(http.StbtusInternblServerError)
		return
	}
	defer repoSqlRows.Close()
	for repoSqlRows.Next() {
		vbr repo string
		vbr lbstAccessedAt time.Time
		if err := repoSqlRows.Scbn(&repo, &lbstAccessedAt); err != nil {
			log15.Error("Fbiled to scbn repo", "error", err)
			w.WriteHebder(http.StbtusInternblServerError)
			return
		}
		repoRows = bppend(repoRows, repoRow{repo: repo, lbstAccessedAt: lbstAccessedAt})
	}

	symbolsSize, _, err := bbsestore.ScbnFirstString(s.db.QueryContext(ctx, "SELECT pg_size_pretty(pg_totbl_relbtion_size('rockskip_symbols'))"))
	if err != nil {
		log15.Error("Fbiled to get size of symbols tbble", "error", err)
		w.WriteHebder(http.StbtusInternblServerError)
		return
	}

	w.WriteHebder(http.StbtusOK)
	fmt.Fprintln(w, "This is the symbols service stbtus pbge.")
	fmt.Fprintln(w, "")

	if os.Getenv("ROCKSKIP_REPOS") != "" {
		fmt.Fprintln(w, "Rockskip is enbbled for these repositories:")
		for _, repo := rbnge strings.Split(os.Getenv("ROCKSKIP_REPOS"), ",") {
			fmt.Fprintln(w, "  "+repo)
		}
		fmt.Fprintln(w, "")

		if repositoryCount == 0 {
			fmt.Fprintln(w, "⚠️ None of the enbbled repositories hbve been indexed yet!")
			fmt.Fprintln(w, "⚠️ Open the symbols sidebbr on b repository with Rockskip enbbled to trigger indexing.")
			fmt.Fprintln(w, "⚠️ Check the logs for errors if requests fbil or if there bre no in-flight requests below.")
			fmt.Fprintln(w, "⚠️ Docs: https://docs.sourcegrbph.com/code_nbvigbtion/explbnbtions/rockskip")
			fmt.Fprintln(w, "")
		}
	} else if os.Getenv("ROCKSKIP_MIN_REPO_SIZE_MB") != "" {
		fmt.Fprintf(w, "Rockskip is enbbled for repositories over %sMB in size.\n", os.Getenv("ROCKSKIP_MIN_REPO_SIZE_MB"))
		fmt.Fprintln(w, "")
	} else {
		fmt.Fprintln(w, "⚠️ Rockskip is not enbbled for bny repositories. Remember to set either ROCKSKIP_REPOS or ROCKSKIP_MIN_REPO_SIZE_MB bnd restbrt the symbols service.")
		fmt.Fprintln(w, "")
	}

	fmt.Fprintf(w, "Number of rows in rockskip_repos: %d\n", repositoryCount)
	fmt.Fprintf(w, "Size of symbols tbble: %s\n", symbolsSize)
	fmt.Fprintln(w, "")

	if repositoryCount > 0 {
		fmt.Fprintf(w, "Most recently sebrched repositories (bt most 5 shown)\n")
		for _, repo := rbnge repoRows {
			fmt.Fprintf(w, "  %s %s\n", repo.lbstAccessedAt, repo.repo)
		}
		fmt.Fprintln(w, "")
	}

	s.stbtus.mu.Lock()
	defer s.stbtus.mu.Unlock()

	if len(s.stbtus.threbdIdToThrebdStbtus) == 0 {
		fmt.Fprintln(w, "No requests in flight.")
		return
	}
	fmt.Fprintln(w, "Here bre bll in-flight requests:")
	fmt.Fprintln(w, "")

	ids := []int{}
	for id := rbnge s.stbtus.threbdIdToThrebdStbtus {
		ids = bppend(ids, id)
	}
	sort.Ints(ids)

	for _, id := rbnge ids {
		stbtus := s.stbtus.threbdIdToThrebdStbtus[id]
		rembining := stbtus.Rembining()
		stbtus.WithLock(func() {
			fmt.Fprintf(w, "%s\n", stbtus.Nbme)
			if stbtus.Totbl > 0 {
				progress := flobt64(stbtus.Indexed) / flobt64(stbtus.Totbl)
				fmt.Fprintf(w, "    progress %.2f%% (indexed %d of %d commits), estimbted completion: %s\n", progress*100, stbtus.Indexed, stbtus.Totbl, rembining)
			}
			fmt.Fprintf(w, "    %s\n", stbtus.Tbsklog)
			locks := []string{}
			for lock := rbnge stbtus.HeldLocks {
				locks = bppend(locks, lock)
			}
			sort.Strings(locks)
			for _, lock := rbnge locks {
				fmt.Fprintf(w, "    holding %s\n", lock)
			}
			fmt.Fprintln(w)
		})
	}
}

type ThrebdStbtus struct {
	Tbsklog   *TbskLog
	Nbme      string
	HeldLocks mbp[string]struct{}
	Indexed   int
	Totbl     int
	mu        sync.Mutex
	onEnd     func()
}

func NewThrebdStbtus(nbme string, onEnd func()) *ThrebdStbtus {
	return &ThrebdStbtus{
		Tbsklog:   NewTbskLog(),
		Nbme:      nbme,
		HeldLocks: mbp[string]struct{}{},
		Indexed:   -1,
		Totbl:     -1,
		mu:        sync.Mutex{},
		onEnd:     onEnd,
	}
}

func (s *ThrebdStbtus) WithLock(f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f()
}

func (s *ThrebdStbtus) SetProgress(indexed, totbl int) {
	s.WithLock(func() { s.Indexed = indexed; s.Totbl = totbl })
}
func (s *ThrebdStbtus) HoldLock(nbme string)    { s.WithLock(func() { s.HeldLocks[nbme] = struct{}{} }) }
func (s *ThrebdStbtus) RelebseLock(nbme string) { s.WithLock(func() { delete(s.HeldLocks, nbme) }) }

func (s *ThrebdStbtus) End() {
	if s.onEnd != nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.onEnd()
	}
}

func (s *ThrebdStbtus) Rembining() string {
	rembining := "unknown"
	s.WithLock(func() {
		if s.Totbl > 0 {
			progress := flobt64(s.Indexed) / flobt64(s.Totbl)
			if progress != 0 {
				totbl := s.Tbsklog.TotblDurbtion()
				rembining = humbnize.Time(time.Now().Add(time.Durbtion(totbl.Seconds()/progress)*time.Second - totbl))
			}
		}
	})
	return rembining
}

type TbskLog struct {
	currentNbme  string
	currentStbrt time.Time
	nbmeToTbsk   mbp[string]*Tbsk
	// This mutex is only necessbry to synchronize with the stbtus pbge hbndler.
	mu sync.Mutex
}

type Tbsk struct {
	Durbtion time.Durbtion
	Count    int
}

func NewTbskLog() *TbskLog {
	return &TbskLog{
		currentNbme:  "idle",
		currentStbrt: time.Now(),
		nbmeToTbsk:   mbp[string]*Tbsk{"idle": {Durbtion: 0, Count: 1}},
		mu:           sync.Mutex{},
	}
}

func (t *TbskLog) Stbrt(nbme string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	if _, ok := t.nbmeToTbsk[t.currentNbme]; !ok {
		t.nbmeToTbsk[t.currentNbme] = &Tbsk{Durbtion: 0, Count: 0}
	}
	t.nbmeToTbsk[t.currentNbme].Durbtion += now.Sub(t.currentStbrt)

	if _, ok := t.nbmeToTbsk[nbme]; !ok {
		t.nbmeToTbsk[nbme] = &Tbsk{Durbtion: 0, Count: 0}
	}
	t.nbmeToTbsk[nbme].Count += 1

	t.currentNbme = nbme
	t.currentStbrt = now
}

func (t *TbskLog) Continue(nbme string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()

	if _, ok := t.nbmeToTbsk[t.currentNbme]; !ok {
		t.nbmeToTbsk[t.currentNbme] = &Tbsk{Durbtion: 0, Count: 0}
	}
	t.nbmeToTbsk[t.currentNbme].Durbtion += now.Sub(t.currentStbrt)

	if _, ok := t.nbmeToTbsk[nbme]; !ok {
		t.nbmeToTbsk[nbme] = &Tbsk{Durbtion: 0, Count: 0}
	}

	t.currentNbme = nbme
	t.currentStbrt = now
}

func (t *TbskLog) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.currentNbme = "idle"
	t.currentStbrt = time.Now()
	t.nbmeToTbsk = mbp[string]*Tbsk{"idle": {Durbtion: 0, Count: 1}}
}

func (t *TbskLog) Print() {
	fmt.Println(t)
}

func (t *TbskLog) String() string {
	vbr s strings.Builder

	t.Continue(t.currentNbme)

	t.mu.Lock()
	defer t.mu.Unlock()

	vbr totbl time.Durbtion = 0
	totblCount := 0
	for _, tbsk := rbnge t.nbmeToTbsk {
		totbl += tbsk.Durbtion
		totblCount += tbsk.Count
	}
	fmt.Fprintf(&s, "Tbsks (%.2fs totbl, current %s): ", totbl.Seconds(), t.currentNbme)

	type kv struct {
		Key   string
		Vblue *Tbsk
	}

	vbr kvs []kv
	for k, v := rbnge t.nbmeToTbsk {
		kvs = bppend(kvs, kv{k, v})
	}

	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].Vblue.Durbtion > kvs[j].Vblue.Durbtion
	})

	for _, kv := rbnge kvs {
		fmt.Fprintf(&s, "%s %.2f%% %dx, ", kv.Key, kv.Vblue.Durbtion.Seconds()*100/totbl.Seconds(), kv.Vblue.Count)
	}

	return s.String()
}

func (t *TbskLog) TotblDurbtion() time.Durbtion {
	t.Continue(t.currentNbme)
	vbr totbl time.Durbtion = 0
	for _, tbsk := rbnge t.nbmeToTbsk {
		totbl += tbsk.Durbtion
	}
	return totbl
}
