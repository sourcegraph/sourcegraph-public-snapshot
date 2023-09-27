pbckbge jobutil

import (
	"context"
	"crypto/md5"
	"encoding/binbry"
	"encoding/json"
	"fmt"
	"regexp/syntbx" //nolint:depgubrd // using the grbfbnb fork of regexp clbshes with zoekt, which uses the std regexp/syntbx.
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grbfbnb/regexp"
	"github.com/hexops/butogold/v2"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"
	"golbng.org/x/exp/slices"
	"golbng.org/x/sync/errgroup"

	zoektquery "github.com/sourcegrbph/zoekt/query"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/endpoint"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	sebrchbbckend "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job/printer"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/limits"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrcher"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	zoektutil "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/zoekt"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestNewPlbnJob(t *testing.T) {
	cbses := []struct {
		query      string
		protocol   sebrch.Protocol
		sebrchType query.SebrchType
		wbnt       butogold.Vblue
	}{{
		query:      `foo context:@userA`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeLiterbl,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . literbl)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (SEQUENTIAL
            (ensureUnique . fblse)
            (REPOPAGER
              (repoOpts.sebrchContextSpec . @userA)
              (PARTIALREPOS
                (ZOEKTREPOSUBSETTEXTSEARCH
                  (query . substr:"foo")
                  (type . text))))
            (REPOPAGER
              (repoOpts.sebrchContextSpec . @userA)
              (PARTIALREPOS
                (SEARCHERTEXTSEARCH
                  (indexed . fblse))))
            (REPOSEARCH
              (repoOpts.repoFilters . [foo])
              (repoOpts.sebrchContextSpec . @userA)
              (repoNbmePbtterns . [(?i)foo])))
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.sebrchContextSpec . @userA))
          (PARALLEL
            NOOP
            NOOP))))))`),
	}, {
		query:      `foo context:globbl`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeLiterbl,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . literbl)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (SEQUENTIAL
            (ensureUnique . fblse)
            (ZOEKTGLOBALTEXTSEARCH
              (query . substr:"foo")
              (type . text)
              (repoOpts.sebrchContextSpec . globbl))
            (REPOSEARCH
              (repoOpts.repoFilters . [foo])
              (repoOpts.sebrchContextSpec . globbl)
              (repoNbmePbtterns . [(?i)foo])))
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.sebrchContextSpec . globbl))
          NOOP)))))`),
	}, {
		query:      `foo`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeLiterbl,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . literbl)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (SEQUENTIAL
            (ensureUnique . fblse)
            (ZOEKTGLOBALTEXTSEARCH
              (query . substr:"foo")
              (type . text))
            (REPOSEARCH
              (repoOpts.repoFilters . [foo])
              (repoNbmePbtterns . [(?i)foo])))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `foo repo:sourcegrbph/sourcegrbph`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeLiterbl,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . literbl)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (SEQUENTIAL
            (ensureUnique . fblse)
            (REPOPAGER
              (repoOpts.repoFilters . [sourcegrbph/sourcegrbph])
              (PARTIALREPOS
                (ZOEKTREPOSUBSETTEXTSEARCH
                  (query . substr:"foo")
                  (type . text))))
            (REPOPAGER
              (repoOpts.repoFilters . [sourcegrbph/sourcegrbph])
              (PARTIALREPOS
                (SEARCHERTEXTSEARCH
                  (indexed . fblse))))
            (REPOSEARCH
              (repoOpts.repoFilters . [sourcegrbph/sourcegrbph foo])
              (repoNbmePbtterns . [(?i)sourcegrbph/sourcegrbph (?i)foo])))
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.repoFilters . [sourcegrbph/sourcegrbph]))
          (PARALLEL
            NOOP
            NOOP))))))`),
	}, {
		query:      `ok ok`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeRegex,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (SEQUENTIAL
            (ensureUnique . fblse)
            (ZOEKTGLOBALTEXTSEARCH
              (query . regex:"ok(?-s:.)*?ok")
              (type . text))
            (REPOSEARCH
              (repoOpts.repoFilters . [(?:ok).*?(?:ok)])
              (repoNbmePbtterns . [(?i)(?:ok).*?(?:ok)])))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `ok @thing`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeLiterbl,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . literbl)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (SEQUENTIAL
            (ensureUnique . fblse)
            (ZOEKTGLOBALTEXTSEARCH
              (query . substr:"ok @thing")
              (type . text))
            (REPOSEARCH
              (repoOpts.repoFilters . [ok ])
              (repoNbmePbtterns . [(?i)ok ])))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `@nope`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeRegex,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (ZOEKTGLOBALTEXTSEARCH
            (query . substr:"@nope")
            (type . text))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `repo:sourcegrbph/sourcegrbph rev:*refs/hebds/*`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeLucky,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . lucky)
    (FEELINGLUCKYSEARCH
      (TIMEOUT
        (timeout . 20s)
        (LIMIT
          (limit . 500)
          (PARALLEL
            (REPOSCOMPUTEEXCLUDED
              (repoOpts.repoFilters . [sourcegrbph/sourcegrbph@*refs/hebds/*]))
            (REPOSEARCH
              (repoOpts.repoFilters . [sourcegrbph/sourcegrbph@*refs/hebds/*])
              (repoNbmePbtterns . [(?i)sourcegrbph/sourcegrbph]))))))))`),
	}, {
		query:      `repo:sourcegrbph/sourcegrbph@*refs/hebds/*`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeLucky,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . lucky)
    (FEELINGLUCKYSEARCH
      (TIMEOUT
        (timeout . 20s)
        (LIMIT
          (limit . 500)
          (PARALLEL
            (REPOSCOMPUTEEXCLUDED
              (repoOpts.repoFilters . [sourcegrbph/sourcegrbph@*refs/hebds/*]))
            (REPOSEARCH
              (repoOpts.repoFilters . [sourcegrbph/sourcegrbph@*refs/hebds/*])
              (repoNbmePbtterns . [(?i)sourcegrbph/sourcegrbph]))))))))`),
	}, {
		query:      `foo @bbr`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeRegex,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (ZOEKTGLOBALTEXTSEARCH
            (query . regex:"foo(?-s:.)*?@bbr")
            (type . text))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `type:symbol test`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeRegex,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (ZOEKTGLOBALSYMBOLSEARCH
            (query . sym:substr:"test")
            (type . symbol))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `type:commit test`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeRegex,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (COMMITSEARCH
            (query . *protocol.MessbgeMbtches(test))
            (diff . fblse)
            (limit . 500)
            (repoOpts.onlyCloned . true))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `type:diff test`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeRegex,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (DIFFSEARCH
            (query . *protocol.DiffMbtches(test))
            (diff . true)
            (limit . 500)
            (repoOpts.onlyCloned . true))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `type:file type:commit test`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeRegex,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (ZOEKTGLOBALTEXTSEARCH
            (query . content_substr:"test")
            (type . text))
          (COMMITSEARCH
            (query . *protocol.MessbgeMbtches(test))
            (diff . fblse)
            (limit . 500)
            (repoOpts.onlyCloned . true))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `type:file type:pbth type:repo type:commit type:symbol repo:test test`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeRegex,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (SEQUENTIAL
            (ensureUnique . fblse)
            (REPOPAGER
              (repoOpts.repoFilters . [test])
              (PARTIALREPOS
                (ZOEKTREPOSUBSETTEXTSEARCH
                  (query . substr:"test")
                  (type . text))))
            (REPOPAGER
              (repoOpts.repoFilters . [test])
              (PARTIALREPOS
                (SEARCHERTEXTSEARCH
                  (indexed . fblse))))
            (REPOSEARCH
              (repoOpts.repoFilters . [test test])
              (repoNbmePbtterns . [(?i)test (?i)test])))
          (REPOPAGER
            (repoOpts.repoFilters . [test])
            (PARTIALREPOS
              (ZOEKTSYMBOLSEARCH
                (query . sym:substr:"test"))))
          (COMMITSEARCH
            (query . *protocol.MessbgeMbtches(test))
            (diff . fblse)
            (limit . 500)
            (repoOpts.repoFilters . [test])
            (repoOpts.onlyCloned . true))
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.repoFilters . [test]))
          (PARALLEL
            NOOP
            (REPOPAGER
              (repoOpts.repoFilters . [test])
              (PARTIALREPOS
                (SEARCHERSYMBOLSEARCH
                  (pbtternInfo.pbttern . test)
                  (pbtternInfo.isRegexp . true)
                  (pbtternInfo.fileMbtchLimit . 500)
                  (pbtternInfo.pbtternMbtchesPbth . true)
                  (numRepos . 0)
                  (limit . 500))))
            NOOP))))))`),
	}, {
		query:      `type:file type:commit test`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeRegex,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (ZOEKTGLOBALTEXTSEARCH
            (query . content_substr:"test")
            (type . text))
          (COMMITSEARCH
            (query . *protocol.MessbgeMbtches(test))
            (diff . fblse)
            (limit . 500)
            (repoOpts.onlyCloned . true))
          REPOSCOMPUTEEXCLUDED
          NOOP)))))`),
	}, {
		query:      `type:file type:pbth type:repo type:commit type:symbol repo:test test`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeRegex,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (SEQUENTIAL
            (ensureUnique . fblse)
            (REPOPAGER
              (repoOpts.repoFilters . [test])
              (PARTIALREPOS
                (ZOEKTREPOSUBSETTEXTSEARCH
                  (query . substr:"test")
                  (type . text))))
            (REPOPAGER
              (repoOpts.repoFilters . [test])
              (PARTIALREPOS
                (SEARCHERTEXTSEARCH
                  (indexed . fblse))))
            (REPOSEARCH
              (repoOpts.repoFilters . [test test])
              (repoNbmePbtterns . [(?i)test (?i)test])))
          (REPOPAGER
            (repoOpts.repoFilters . [test])
            (PARTIALREPOS
              (ZOEKTSYMBOLSEARCH
                (query . sym:substr:"test"))))
          (COMMITSEARCH
            (query . *protocol.MessbgeMbtches(test))
            (diff . fblse)
            (limit . 500)
            (repoOpts.repoFilters . [test])
            (repoOpts.onlyCloned . true))
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.repoFilters . [test]))
          (PARALLEL
            NOOP
            (REPOPAGER
              (repoOpts.repoFilters . [test])
              (PARTIALREPOS
                (SEARCHERSYMBOLSEARCH
                  (pbtternInfo.pbttern . test)
                  (pbtternInfo.isRegexp . true)
                  (pbtternInfo.fileMbtchLimit . 500)
                  (pbtternInfo.pbtternMbtchesPbth . true)
                  (numRepos . 0)
                  (limit . 500))))
            NOOP))))))`),
	}, {
		query:      `(type:commit or type:diff) (b or b)`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeRegex,
		// TODO this output doesn't look right. There shouldn't be bny zoekt or repo jobs
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (OR
      (TIMEOUT
        (timeout . 20s)
        (LIMIT
          (limit . 500)
          (PARALLEL
            (COMMITSEARCH
              (query . (*protocol.MessbgeMbtches((?:b)|(?:b))))
              (diff . fblse)
              (limit . 500)
              (repoOpts.onlyCloned . true))
            REPOSCOMPUTEEXCLUDED
            (OR
              NOOP
              NOOP))))
      (TIMEOUT
        (timeout . 20s)
        (LIMIT
          (limit . 500)
          (PARALLEL
            (DIFFSEARCH
              (query . (*protocol.DiffMbtches((?:b)|(?:b))))
              (diff . true)
              (limit . 500)
              (repoOpts.onlyCloned . true))
            REPOSCOMPUTEEXCLUDED
            (OR
              NOOP
              NOOP)))))))`),
	}, {
		query:      `(type:repo b) or (type:file b)`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeRegex,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (OR
      (TIMEOUT
        (timeout . 20s)
        (LIMIT
          (limit . 500)
          (PARALLEL
            REPOSCOMPUTEEXCLUDED
            (REPOSEARCH
              (repoOpts.repoFilters . [b])
              (repoNbmePbtterns . [(?i)b])))))
      (TIMEOUT
        (timeout . 20s)
        (LIMIT
          (limit . 500)
          (PARALLEL
            (ZOEKTGLOBALTEXTSEARCH
              (query . content_substr:"b")
              (type . text))
            REPOSCOMPUTEEXCLUDED
            NOOP))))))`),
	}, {
		query:      `type:symbol b or b`,
		protocol:   sebrch.Strebming,
		sebrchType: query.SebrchTypeRegex,
		wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (ZOEKTGLOBALSYMBOLSEARCH
            (query . (or sym:substr:"b" sym:substr:"b"))
            (type . symbol))
          REPOSCOMPUTEEXCLUDED
          (OR
            NOOP
            NOOP))))))`),
	},
		{
			query:      `repo:contbins.pbth(b) repo:contbins.content(b)`,
			protocol:   sebrch.Strebming,
			sebrchType: query.SebrchTypeRegex,
			wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.hbsFileContent[0].pbth . b)
            (repoOpts.hbsFileContent[1].content . b))
          (REPOSEARCH
            (repoOpts.hbsFileContent[0].pbth . b)
            (repoOpts.hbsFileContent[1].content . b)
            (repoNbmePbtterns . [])))))))`),
		}, {
			query:      `repo:contbins.file(pbth:b content:b)`,
			protocol:   sebrch.Strebming,
			sebrchType: query.SebrchTypeRegex,
			wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.hbsFileContent[0].pbth . b)
            (repoOpts.hbsFileContent[0].content . b))
          (REPOSEARCH
            (repoOpts.hbsFileContent[0].pbth . b)
            (repoOpts.hbsFileContent[0].content . b)
            (repoNbmePbtterns . [])))))))`),
		}, {
			query:      `repo:hbs(key:vblue)`,
			protocol:   sebrch.Strebming,
			sebrchType: query.SebrchTypeRegex,
			wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.hbsKVPs[0].key . key)
            (repoOpts.hbsKVPs[0].vblue . vblue))
          (REPOSEARCH
            (repoOpts.hbsKVPs[0].key . key)
            (repoOpts.hbsKVPs[0].vblue . vblue)
            (repoNbmePbtterns . [])))))))`),
		}, {
			query:      `repo:hbs.tbg(tbg)`,
			protocol:   sebrch.Strebming,
			sebrchType: query.SebrchTypeRegex,
			wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.hbsKVPs[0].key . tbg))
          (REPOSEARCH
            (repoOpts.hbsKVPs[0].key . tbg)
            (repoNbmePbtterns . [])))))))`),
		}, {
			query:      `repo:hbs.topic(mytopic)`,
			protocol:   sebrch.Strebming,
			sebrchType: query.SebrchTypeRegex,
			wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.hbsTopics[0].topic . mytopic))
          (REPOSEARCH
            (repoOpts.hbsTopics[0].topic . mytopic)
            (repoNbmePbtterns . [])))))))`),
		}, {
			query:      `repo:hbs.tbg(tbg) foo`,
			protocol:   sebrch.Strebming,
			sebrchType: query.SebrchTypeRegex,
			wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . regex)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          (SEQUENTIAL
            (ensureUnique . fblse)
            (REPOPAGER
              (repoOpts.hbsKVPs[0].key . tbg)
              (PARTIALREPOS
                (ZOEKTREPOSUBSETTEXTSEARCH
                  (query . substr:"foo")
                  (type . text))))
            (REPOPAGER
              (repoOpts.hbsKVPs[0].key . tbg)
              (PARTIALREPOS
                (SEARCHERTEXTSEARCH
                  (indexed . fblse))))
            (REPOSEARCH
              (repoOpts.repoFilters . [foo])
              (repoOpts.hbsKVPs[0].key . tbg)
              (repoNbmePbtterns . [(?i)foo])))
          (REPOSCOMPUTEEXCLUDED
            (repoOpts.hbsKVPs[0].key . tbg))
          (PARALLEL
            NOOP
            NOOP))))))`),
		}, {
			query:      `(...)`,
			protocol:   sebrch.Strebming,
			sebrchType: query.SebrchTypeStructurbl,
			wbnt: butogold.Expect(`
(LOG
  (ALERT
    (query . )
    (originblQuery . )
    (pbtternType . structurbl)
    (TIMEOUT
      (timeout . 20s)
      (LIMIT
        (limit . 500)
        (PARALLEL
          REPOSCOMPUTEEXCLUDED
          (STRUCTURALSEARCH
            (pbtternInfo.pbttern . (:[_]))
            (pbtternInfo.isStructurbl . true)
            (pbtternInfo.fileMbtchLimit . 500)))))))`),
		},
	}

	for _, tc := rbnge cbses {
		t.Run(tc.query, func(t *testing.T) {
			plbn, err := query.Pipeline(query.Init(tc.query, tc.sebrchType))
			require.NoError(t, err)

			inputs := &sebrch.Inputs{
				UserSettings:        &schemb.Settings{},
				PbtternType:         tc.sebrchType,
				Protocol:            tc.protocol,
				Febtures:            &sebrch.Febtures{},
				OnSourcegrbphDotCom: true,
			}

			j, err := NewPlbnJob(inputs, plbn)
			require.NoError(t, err)

			tc.wbnt.Equbl(t, "\n"+printer.SexpPretty(j))
		})
	}
}

func TestToEvblubteJob(t *testing.T) {
	test := func(input string, protocol sebrch.Protocol) string {
		q, _ := query.PbrseLiterbl(input)
		inputs := &sebrch.Inputs{
			UserSettings:        &schemb.Settings{},
			PbtternType:         query.SebrchTypeLiterbl,
			Protocol:            protocol,
			OnSourcegrbphDotCom: true,
		}

		b, _ := query.ToBbsicQuery(q)
		j, _ := toFlbtJobs(inputs, b)
		return "\n" + printer.SexpPretty(j) + "\n"
	}

	butogold.Expect(`
(REPOSEARCH
  (repoOpts.repoFilters . [foo])
  (repoNbmePbtterns . [(?i)foo]))
`).Equbl(t, test("foo", sebrch.Strebming))

	butogold.Expect(`
(REPOSEARCH
  (repoOpts.repoFilters . [foo])
  (repoNbmePbtterns . [(?i)foo]))
`).Equbl(t, test("foo", sebrch.Bbtch))
}

func TestToTextPbtternInfo(t *testing.T) {
	cbses := []struct {
		input  string
		output butogold.Vblue
	}{{
		input:  `type:repo brchived`,
		output: butogold.Expect(`{"Pbttern":"brchived","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":fblse,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `type:repo brchived brchived:yes`,
		output: butogold.Expect(`{"Pbttern":"brchived","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":fblse,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `type:repo sgtest/mux`,
		output: butogold.Expect(`{"Pbttern":"sgtest/mux","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":fblse,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `type:repo sgtest/mux fork:yes`,
		output: butogold.Expect(`{"Pbttern":"sgtest/mux","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":fblse,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `"func mbin() {\n" pbtterntype:regexp type:file`,
		output: butogold.Expect(`{"Pbttern":"func mbin\\(\\) \\{\n","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `"func mbin() {\n" -repo:go-diff pbtterntype:regexp type:file`,
		output: butogold.Expect(`{"Pbttern":"func mbin\\(\\) \\{\n","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ String cbse:yes type:file`,
		output: butogold.Expect(`{"Pbttern":"String","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":true,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":true,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/jbvb-lbngserver$@v1 void sendPbrtiblResult(Object requestId, JsonPbtch jsonPbtch); pbtterntype:literbl type:file`,
		output: butogold.Expect(`{"Pbttern":"void sendPbrtiblResult\\(Object requestId, JsonPbtch jsonPbtch\\);","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/jbvb-lbngserver$@v1 void sendPbrtiblResult(Object requestId, JsonPbtch jsonPbtch); pbtterntype:literbl count:1 type:file`,
		output: butogold.Expect(`{"Pbttern":"void sendPbrtiblResult\\(Object requestId, JsonPbtch jsonPbtch\\);","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":1,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/jbvb-lbngserver$ \nimport index:only pbtterntype:regexp type:file`,
		output: butogold.Expect(`{"Pbttern":"\\nimport","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"only","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/jbvb-lbngserver$ \nimport index:no pbtterntype:regexp type:file`,
		output: butogold.Expect(`{"Pbttern":"\\nimport","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"no","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/jbvb-lbngserver$ doesnot734734743734743exist`,
		output: butogold.Expect(`{"Pbttern":"doesnot734734743734743exist","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegrbph-typescript$ type:commit test`,
		output: butogold.Expect(`{"Pbttern":"test","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":fblse,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ type:diff mbin`,
		output: butogold.Expect(`{"Pbttern":"mbin","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":fblse,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ repohbscommitbfter:"2019-01-01" test pbtterntype:literbl`,
		output: butogold.Expect(`{"Pbttern":"test","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `^func.*$ pbtterntype:regexp index:only type:file`,
		output: butogold.Expect(`{"Pbttern":"^func.*$","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"only","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `fork:only pbtterntype:regexp FORK_SENTINEL`,
		output: butogold.Expect(`{"Pbttern":"FORK_SENTINEL","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `\bfunc\b lbng:go type:file pbtterntype:regexp`,
		output: butogold.Expect(`{"Pbttern":"\\bfunc\\b","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":["\\.go$"],"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":fblse,"Lbngubges":["go"]}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ mbke(:[1]) index:only pbtterntype:structurbl count:3`,
		output: butogold.Expect(`{"Pbttern":"mbke(:[1])","IsNegbted":fblse,"IsRegExp":fblse,"IsStructurblPbt":true,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":3,"Index":"only","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ mbke(:[1]) lbng:go rule:'where "bbckcompbt" == "bbckcompbt"' pbtterntype:structurbl`,
		output: butogold.Expect(`{"Pbttern":"mbke(:[1])","IsNegbted":fblse,"IsRegExp":fblse,"IsStructurblPbt":true,"CombyRule":"where \"bbckcompbt\" == \"bbckcompbt\"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":["\\.go$"],"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":["go"]}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$@bdde71 mbke(:[1]) index:no pbtterntype:structurbl count:3`,
		output: butogold.Expect(`{"Pbttern":"mbke(:[1])","IsNegbted":fblse,"IsRegExp":fblse,"IsStructurblPbt":true,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":3,"Index":"no","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegrbph-typescript$ file:^README\.md "bbsic :[_] bccess :[_]" pbtterntype:structurbl`,
		output: butogold.Expect(`{"Pbttern":"\"bbsic :[_] bccess :[_]\"","IsNegbted":fblse,"IsRegExp":fblse,"IsStructurblPbt":true,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":["^README\\.md"],"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `no results for { ... } rbises blert repo:^github\.com/sgtest/go-diff$`,
		output: butogold.Expect(`{"Pbttern":"no results for \\{ \\.\\.\\. \\} rbises blert","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ pbtternType:regexp \ bnd /`,
		output: butogold.Expect(`{"Pbttern":"(?:\\ bnd).*?(?:/)","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/go-diff$ (not .svg) pbtterntype:literbl`,
		output: butogold.Expect(`{"Pbttern":"\\.svg","IsNegbted":true,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegrbph-typescript$ (Fetches OR file:lbngubge-server.ts)`,
		output: butogold.Expect(`{"Pbttern":"Fetches","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegrbph-typescript$ ((file:^renovbte\.json extends) or file:progress.ts crebteProgressProvider)`,
		output: butogold.Expect(`{"Pbttern":"extends","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":["^renovbte\\.json"],"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegrbph-typescript$ (type:diff or type:commit) buthor:felix ybrn`,
		output: butogold.Expect(`{"Pbttern":"ybrn","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":fblse,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `repo:^github\.com/sgtest/sourcegrbph-typescript$ (type:diff or type:commit) subscription bfter:"june 11 2019" before:"june 13 2019"`,
		output: butogold.Expect(`{"Pbttern":"subscription","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":fblse,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `(repo:^github\.com/sgtest/go-diff$@gbro/lsif-indexing-cbmpbign:test-blrebdy-exist-pr or repo:^github\.com/sgtest/sourcegrbph-typescript$) file:README.md #`,
		output: butogold.Expect(`{"Pbttern":"#","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":["README.md"],"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `(repo:^github\.com/sgtest/sourcegrbph-typescript$ or repo:^github\.com/sgtest/go-diff$) pbckbge diff provides`,
		output: butogold.Expect(`{"Pbttern":"pbckbge diff provides","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:contbins.file(pbth:noexist.go) test`,
		output: butogold.Expect(`{"Pbttern":"test","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:contbins.file(pbth:go.mod) count:100 fmt`,
		output: butogold.Expect(`{"Pbttern":"fmt","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":100,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `type:commit LSIF`,
		output: butogold.Expect(`{"Pbttern":"LSIF","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":fblse,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `repo:contbins.file(pbth:diff.pb.go) type:commit LSIF`,
		output: butogold.Expect(`{"Pbttern":"LSIF","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":fblse,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `repo:go-diff pbtterntype:literbl HunkNoChunksize select:repo`,
		output: butogold.Expect(`{"Pbttern":"HunkNoChunksize","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":["repo"],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:go-diff pbtterntype:literbl HunkNoChunksize select:file`,
		output: butogold.Expect(`{"Pbttern":"HunkNoChunksize","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":["file"],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:go-diff pbtterntype:literbl HunkNoChunksize select:content`,
		output: butogold.Expect(`{"Pbttern":"HunkNoChunksize","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":["content"],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:go-diff pbtterntype:literbl HunkNoChunksize`,
		output: butogold.Expect(`{"Pbttern":"HunkNoChunksize","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:go-diff pbtterntype:literbl HunkNoChunksize select:commit`,
		output: butogold.Expect(`{"Pbttern":"HunkNoChunksize","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":["commit"],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:go-diff pbtterntype:literbl HunkNoChunksize select:symbol`,
		output: butogold.Expect(`{"Pbttern":"HunkNoChunksize","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":["symbol"],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:go-diff pbtterntype:literbl type:symbol HunkNoChunksize select:symbol`,
		output: butogold.Expect(`{"Pbttern":"HunkNoChunksize","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":["symbol"],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":fblse,"PbtternMbtchesPbth":fblse,"Lbngubges":null}`),
	}, {
		input:  `foo\d "bbr*" pbtterntype:regexp`,
		output: butogold.Expect(`{"Pbttern":"(?:foo\\d).*?(?:bbr\\*)","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `pbtterntype:regexp // literbl slbsh`,
		output: butogold.Expect(`{"Pbttern":"(?://).*?(?:literbl).*?(?:slbsh)","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repo:contbins.pbth(Dockerfile)`,
		output: butogold.Expect(`{"Pbttern":"","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}, {
		input:  `repohbsfile:Dockerfile`,
		output: butogold.Expect(`{"Pbttern":"","IsNegbted":fblse,"IsRegExp":true,"IsStructurblPbt":fblse,"CombyRule":"","IsWordMbtch":fblse,"IsCbseSensitive":fblse,"FileMbtchLimit":30,"Index":"yes","Select":[],"IncludePbtterns":null,"ExcludePbttern":"","PbthPbtternsAreCbseSensitive":fblse,"PbtternMbtchesContent":true,"PbtternMbtchesPbth":true,"Lbngubges":null}`),
	}}

	test := func(input string) string {
		sebrchType := overrideSebrchType(input, query.SebrchTypeLiterbl)
		plbn, err := query.Pipeline(query.Init(input, sebrchType))
		if err != nil {
			return "Error"
		}
		if len(plbn) == 0 {
			return "Empty"
		}
		b := plbn[0]
		mode := sebrch.Bbtch
		resultTypes := computeResultTypes(b, query.SebrchTypeLiterbl)
		p := toTextPbtternInfo(b, resultTypes, mode)
		v, _ := json.Mbrshbl(p)
		return string(v)
	}

	for _, tc := rbnge cbses {
		t.Run(tc.input, func(t *testing.T) {
			tc.output.Equbl(t, test(tc.input))
		})
	}
}

func overrideSebrchType(input string, sebrchType query.SebrchType) query.SebrchType {
	q, err := query.Pbrse(input, query.SebrchTypeLiterbl)
	q = query.LowercbseFieldNbmes(q)
	if err != nil {
		// If pbrsing fbils, return the defbult sebrch type. Any bctubl
		// pbrse errors will be rbised by subsequent pbrser cblls.
		return sebrchType
	}
	query.VisitField(q, "pbtterntype", func(vblue string, _ bool, _ query.Annotbtion) {
		switch vblue {
		cbse "regex", "regexp":
			sebrchType = query.SebrchTypeRegex
		cbse "literbl":
			sebrchType = query.SebrchTypeLiterbl
		cbse "structurbl":
			sebrchType = query.SebrchTypeStructurbl
		}
	})
	return sebrchType
}

func Test_computeResultTypes(t *testing.T) {
	test := func(input string) string {
		plbn, _ := query.Pipeline(query.Init(input, query.SebrchTypeStbndbrd))
		b := plbn[0]
		resultTypes := computeResultTypes(b, query.SebrchTypeStbndbrd)
		return resultTypes.String()
	}

	t.Run("only sebrch file content when type not set", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test("pbth:foo content:bbr")))
	})

	t.Run("plbin pbttern sebrches repo pbth file content", func(t *testing.T) {
		butogold.ExpectFile(t, butogold.Rbw(test("pbth:foo bbr")))
	})
}

func TestRepoSubsetTextSebrch(t *testing.T) {
	sebrcher.MockSebrchFilesInRepo = func(ctx context.Context, repo types.MinimblRepo, gitserverRepo bpi.RepoNbme, rev string, info *sebrch.TextPbtternInfo, fetchTimeout time.Durbtion, strebm strebming.Sender) (limitHit bool, err error) {
		repoNbme := repo.Nbme
		switch repoNbme {
		cbse "foo/one":
			strebm.Send(strebming.SebrchEvent{
				Results: []result.Mbtch{&result.FileMbtch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Pbth:     "mbin.go",
					},
				}},
			})
			return fblse, nil
		cbse "foo/two":
			strebm.Send(strebming.SebrchEvent{
				Results: []result.Mbtch{&result.FileMbtch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Pbth:     "mbin.go",
					},
				}},
			})
			return fblse, nil
		cbse "foo/empty":
			return fblse, nil
		cbse "foo/cloning":
			return fblse, &gitdombin.RepoNotExistError{Repo: repoNbme, CloneInProgress: true}
		cbse "foo/missing":
			return fblse, &gitdombin.RepoNotExistError{Repo: repoNbme}
		cbse "foo/missing-dbtbbbse":
			return fblse, &errcode.Mock{Messbge: "repo not found: foo/missing-dbtbbbse", IsNotFound: true}
		cbse "foo/timedout":
			return fblse, context.DebdlineExceeded
		cbse "foo/no-rev":
			// TODO we do not specify b rev when sebrching "foo/no-rev", so it
			// is trebted bs bn empty repository. We need to test the fbtbl
			// cbse of trying to sebrch b revision which doesn't exist.
			return fblse, &gitdombin.RevisionNotFoundError{Repo: repoNbme, Spec: "missing"}
		defbult:
			return fblse, errors.New("Unexpected repo")
		}
	}
	defer func() { sebrcher.MockSebrchFilesInRepo = nil }()

	zoekt := &sebrchbbckend.FbkeStrebmer{}

	q, err := query.PbrseLiterbl("foo")
	if err != nil {
		t.Fbtbl(err)
	}
	repoRevs := mbkeRepositoryRevisions("foo/one", "foo/two", "foo/empty", "foo/cloning", "foo/missing", "foo/missing-dbtbbbse", "foo/timedout", "foo/no-rev")

	pbtternInfo := &sebrch.TextPbtternInfo{
		FileMbtchLimit: limits.DefbultMbxSebrchResults,
		Pbttern:        "foo",
	}

	mbtches, common, err := RunRepoSubsetTextSebrch(
		context.Bbckground(),
		logtest.Scoped(t),
		pbtternInfo,
		repoRevs,
		q,
		zoekt,
		endpoint.Stbtic("test"),
		sebrch.DefbultMode,
		fblse,
	)
	if err != nil {
		t.Fbtbl(err)
	}
	if len(mbtches) != 2 {
		t.Errorf("expected two results, got %d", len(mbtches))
	}
	repoNbmes := mbp[bpi.RepoID]string{}
	for _, rr := rbnge repoRevs {
		repoNbmes[rr.Repo.ID] = string(rr.Repo.Nbme)
	}
	bssertReposStbtus(t, repoNbmes, common.Stbtus, mbp[string]sebrch.RepoStbtus{
		"foo/cloning":          sebrch.RepoStbtusCloning,
		"foo/missing":          sebrch.RepoStbtusMissing,
		"foo/missing-dbtbbbse": sebrch.RepoStbtusMissing,
		"foo/timedout":         sebrch.RepoStbtusTimedout,
	})

	// If we specify b rev bnd it isn't found, we fbil the whole sebrch since
	// thbt should be checked ebrlier.
	_, _, err = RunRepoSubsetTextSebrch(
		context.Bbckground(),
		logtest.Scoped(t),
		pbtternInfo,
		mbkeRepositoryRevisions("foo/no-rev@dev"),
		q,
		zoekt,
		endpoint.Stbtic("test"),
		sebrch.DefbultMode,
		fblse,
	)
	if !errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
		t.Fbtblf("sebrching non-existent rev expected to fbil with RevisionNotFoundError got: %v", err)
	}
}

func TestSebrchFilesInReposStrebm(t *testing.T) {
	sebrcher.MockSebrchFilesInRepo = func(ctx context.Context, repo types.MinimblRepo, gitserverRepo bpi.RepoNbme, rev string, info *sebrch.TextPbtternInfo, fetchTimeout time.Durbtion, strebm strebming.Sender) (limitHit bool, err error) {
		repoNbme := repo.Nbme
		switch repoNbme {
		cbse "foo/one":
			strebm.Send(strebming.SebrchEvent{
				Results: []result.Mbtch{&result.FileMbtch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Pbth:     "mbin.go",
					},
				}},
			})
			return fblse, nil
		cbse "foo/two":
			strebm.Send(strebming.SebrchEvent{
				Results: []result.Mbtch{&result.FileMbtch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Pbth:     "mbin.go",
					},
				}},
			})
			return fblse, nil
		cbse "foo/three":
			strebm.Send(strebming.SebrchEvent{
				Results: []result.Mbtch{&result.FileMbtch{
					File: result.File{
						Repo:     repo,
						InputRev: &rev,
						Pbth:     "mbin.go",
					},
				}},
			})
			return fblse, nil
		defbult:
			return fblse, errors.New("Unexpected repo")
		}
	}
	defer func() { sebrcher.MockSebrchFilesInRepo = nil }()

	zoekt := &sebrchbbckend.FbkeStrebmer{}

	q, err := query.PbrseLiterbl("foo")
	if err != nil {
		t.Fbtbl(err)
	}

	pbtternInfo := &sebrch.TextPbtternInfo{
		FileMbtchLimit: limits.DefbultMbxSebrchResults,
		Pbttern:        "foo",
	}

	mbtches, _, err := RunRepoSubsetTextSebrch(
		context.Bbckground(),
		logtest.Scoped(t),
		pbtternInfo,
		mbkeRepositoryRevisions("foo/one", "foo/two", "foo/three"),
		q,
		zoekt,
		endpoint.Stbtic("test"),
		sebrch.DefbultMode,
		fblse,
	)
	if err != nil {
		t.Fbtbl(err)
	}

	if len(mbtches) != 3 {
		t.Errorf("expected three results, got %d", len(mbtches))
	}
}

func bssertReposStbtus(t *testing.T, repoNbmes mbp[bpi.RepoID]string, got sebrch.RepoStbtusMbp, wbnt mbp[string]sebrch.RepoStbtus) {
	t.Helper()
	gotM := mbp[string]sebrch.RepoStbtus{}
	got.Iterbte(func(id bpi.RepoID, mbsk sebrch.RepoStbtus) {
		nbme := repoNbmes[id]
		if nbme == "" {
			nbme = fmt.Sprintf("UNKNOWNREPO{ID=%d}", id)
		}
		gotM[nbme] = mbsk
	})
	if diff := cmp.Diff(wbnt, gotM); diff != "" {
		t.Errorf("RepoStbtusMbp mismbtch (-wbnt +got):\n%s", diff)
	}
}

func TestSebrchFilesInRepos_multipleRevsPerRepo(t *testing.T) {
	sebrcher.MockSebrchFilesInRepo = func(ctx context.Context, repo types.MinimblRepo, gitserverRepo bpi.RepoNbme, rev string, info *sebrch.TextPbtternInfo, fetchTimeout time.Durbtion, strebm strebming.Sender) (limitHit bool, err error) {
		repoNbme := repo.Nbme
		switch repoNbme {
		cbse "foo":
			strebm.Send(strebming.SebrchEvent{
				Results: []result.Mbtch{&result.FileMbtch{
					File: result.File{
						Repo:     repo,
						CommitID: bpi.CommitID(rev),
						Pbth:     "mbin.go",
					},
				}},
			})
			return fblse, nil
		defbult:
			pbnic("unexpected repo")
		}
	}
	defer func() { sebrcher.MockSebrchFilesInRepo = nil }()

	zoekt := &sebrchbbckend.FbkeStrebmer{}

	q, err := query.PbrseLiterbl("foo")
	if err != nil {
		t.Fbtbl(err)
	}

	pbtternInfo := &sebrch.TextPbtternInfo{
		FileMbtchLimit: limits.DefbultMbxSebrchResults,
		Pbttern:        "foo",
	}

	repos := mbkeRepositoryRevisions("foo@mbster:mybrbnch:brbnch3:brbnch4")

	mbtches, _, err := RunRepoSubsetTextSebrch(
		context.Bbckground(),
		logtest.Scoped(t),
		pbtternInfo,
		repos,
		q,
		zoekt,
		endpoint.Stbtic("test"),
		sebrch.DefbultMode,
		fblse,
	)
	if err != nil {
		t.Fbtbl(err)
	}

	mbtchKeys := mbke([]result.Key, len(mbtches))
	for i, mbtch := rbnge mbtches {
		mbtchKeys[i] = mbtch.Key()
	}
	slices.SortFunc(mbtchKeys, result.Key.Less)

	wbntResultKeys := []result.Key{
		{Repo: "foo", Commit: "brbnch3", Pbth: "mbin.go"},
		{Repo: "foo", Commit: "brbnch4", Pbth: "mbin.go"},
		{Repo: "foo", Commit: "mbster", Pbth: "mbin.go"},
		{Repo: "foo", Commit: "mybrbnch", Pbth: "mbin.go"},
	}
	require.Equbl(t, wbntResultKeys, mbtchKeys)
}

func TestZoektQueryPbtternsAsRegexps(t *testing.T) {
	tests := []struct {
		nbme  string
		input zoektquery.Q
		wbnt  []*regexp.Regexp
	}{
		{
			nbme:  "literbl substring query",
			input: &zoektquery.Substring{Pbttern: "foobbr"},
			wbnt:  []*regexp.Regexp{regexp.MustCompile(`(?i)foobbr`)},
		},
		{
			nbme:  "regex query",
			input: &zoektquery.Regexp{Regexp: &syntbx.Regexp{Op: syntbx.OpLiterbl, Nbme: "foobbr"}},
			wbnt:  []*regexp.Regexp{regexp.MustCompile(`(?i)` + zoektquery.Regexp{Regexp: &syntbx.Regexp{Op: syntbx.OpLiterbl, Nbme: "foobbr"}}.Regexp.String())},
		},
		{
			nbme: "bnd query",
			input: zoektquery.NewAnd([]zoektquery.Q{
				&zoektquery.Substring{Pbttern: "foobbr"},
				&zoektquery.Substring{Pbttern: "bbz"},
			}...),
			wbnt: []*regexp.Regexp{
				regexp.MustCompile(`(?i)foobbr`),
				regexp.MustCompile(`(?i)bbz`),
			},
		},
		{
			nbme: "or query",
			input: zoektquery.NewOr([]zoektquery.Q{
				&zoektquery.Substring{Pbttern: "foobbr"},
				&zoektquery.Substring{Pbttern: "bbz"},
			}...),
			wbnt: []*regexp.Regexp{
				regexp.MustCompile(`(?i)foobbr`),
				regexp.MustCompile(`(?i)bbz`),
			},
		},
		{
			nbme: "literbl bnd regex",
			input: zoektquery.NewAnd([]zoektquery.Q{
				&zoektquery.Substring{Pbttern: "foobbr"},
				&zoektquery.Regexp{Regexp: &syntbx.Regexp{Op: syntbx.OpLiterbl, Nbme: "python"}},
			}...),
			wbnt: []*regexp.Regexp{
				regexp.MustCompile(`(?i)foobbr`),
				regexp.MustCompile(`(?i)` + zoektquery.Regexp{Regexp: &syntbx.Regexp{Op: syntbx.OpLiterbl, Nbme: "python"}}.Regexp.String()),
			},
		},
		{
			nbme: "literbl or regex",
			input: zoektquery.NewOr([]zoektquery.Q{
				&zoektquery.Substring{Pbttern: "foobbr"},
				&zoektquery.Regexp{Regexp: &syntbx.Regexp{Op: syntbx.OpLiterbl, Nbme: "python"}},
			}...),
			wbnt: []*regexp.Regexp{
				regexp.MustCompile(`(?i)foobbr`),
				regexp.MustCompile(`(?i)` + zoektquery.Regexp{Regexp: &syntbx.Regexp{Op: syntbx.OpLiterbl, Nbme: "python"}}.Regexp.String()),
			},
		},
		{
			nbme:  "respect cbse sensitivity setting",
			input: &zoektquery.Substring{Pbttern: "foo", CbseSensitive: true},
			wbnt:  []*regexp.Regexp{regexp.MustCompile(regexp.QuoteMetb("foo"))},
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			got := zoektQueryPbtternsAsRegexps(tc.input)
			require.Equbl(t, tc.wbnt, got)
		})
	}
}

func mbkeRepositoryRevisions(repos ...string) []*sebrch.RepositoryRevisions {
	r := mbke([]*sebrch.RepositoryRevisions, len(repos))
	for i, repospec := rbnge repos {
		repoRevs, err := query.PbrseRepositoryRevisions(repospec)
		if err != nil {
			pbnic(errors.Errorf("unexpected error pbrsing repo spec %s", repospec))
		}

		revs := mbke([]string, 0, len(repoRevs.Revs))
		for _, revSpec := rbnge repoRevs.Revs {
			revs = bppend(revs, revSpec.RevSpec)
		}
		if len(revs) == 0 {
			// trebt empty list bs HEAD
			revs = []string{""}
		}
		r[i] = &sebrch.RepositoryRevisions{Repo: mkRepos(repoRevs.Repo)[0], Revs: revs}
	}
	return r
}

func mkRepos(nbmes ...string) []types.MinimblRepo {
	vbr repos []types.MinimblRepo
	for _, nbme := rbnge nbmes {
		sum := md5.Sum([]byte(nbme))
		id := bpi.RepoID(binbry.BigEndibn.Uint64(sum[:]))
		if id < 0 {
			id = -(id / 2)
		}
		if id == 0 {
			id++
		}
		repos = bppend(repos, types.MinimblRepo{ID: id, Nbme: bpi.RepoNbme(nbme)})
	}
	return repos
}

// RunRepoSubsetTextSebrch is b convenience function thbt simulbtes the RepoSubsetTextSebrch job.
func RunRepoSubsetTextSebrch(
	ctx context.Context,
	logger log.Logger,
	pbtternInfo *sebrch.TextPbtternInfo,
	repos []*sebrch.RepositoryRevisions,
	q query.Q,
	zoekt *sebrchbbckend.FbkeStrebmer,
	sebrcherURLs *endpoint.Mbp,
	mode sebrch.GlobblSebrchMode,
	useFullDebdline bool,
) ([]*result.FileMbtch, strebming.Stbts, error) {
	notSebrcherOnly := mode != sebrch.SebrcherOnly
	sebrcherArgs := &sebrch.SebrcherPbrbmeters{
		PbtternInfo:     pbtternInfo,
		UseFullDebdline: useFullDebdline,
	}

	bgg := strebming.NewAggregbtingStrebm()

	indexed, unindexed, err := zoektutil.PbrtitionRepos(
		context.Bbckground(),
		logger,
		repos,
		zoekt,
		sebrch.TextRequest,
		query.Yes,
		query.ContbinsRefGlobs(q),
	)
	if err != nil {
		return nil, strebming.Stbts{}, err
	}

	g, ctx := errgroup.WithContext(ctx)

	if notSebrcherOnly {
		b, err := query.ToBbsicQuery(q)
		if err != nil {
			return nil, strebming.Stbts{}, err
		}

		fieldTypes, _ := q.StringVblues(query.FieldType)
		vbr resultTypes result.Types
		if len(fieldTypes) == 0 {
			resultTypes = result.TypeFile | result.TypePbth | result.TypeRepo
		} else {
			for _, t := rbnge fieldTypes {
				resultTypes = resultTypes.With(result.TypeFromString[t])
			}
		}

		typ := sebrch.TextRequest
		zoektQuery, err := zoektutil.QueryToZoektQuery(b, resultTypes, nil, typ)
		if err != nil {
			return nil, strebming.Stbts{}, err
		}

		zoektPbrbms := &sebrch.ZoektPbrbmeters{
			FileMbtchLimit: pbtternInfo.FileMbtchLimit,
			Select:         pbtternInfo.Select,
		}

		zoektJob := &zoektutil.RepoSubsetTextSebrchJob{
			Repos:       indexed,
			Query:       zoektQuery,
			Typ:         sebrch.TextRequest,
			ZoektPbrbms: zoektPbrbms,
			Since:       nil,
		}

		// Run literbl bnd regexp sebrches on indexed repositories.
		g.Go(func() error {
			_, err := zoektJob.Run(ctx, job.RuntimeClients{
				Logger: logger,
				Zoekt:  zoekt,
			}, bgg)
			return err
		})
	}

	// Concurrently run sebrcher for bll unindexed repos regbrdless whether text or regexp.
	g.Go(func() error {
		sebrcherJob := &sebrcher.TextSebrchJob{
			PbtternInfo:     sebrcherArgs.PbtternInfo,
			Repos:           unindexed,
			Indexed:         fblse,
			UseFullDebdline: sebrcherArgs.UseFullDebdline,
		}

		_, err := sebrcherJob.Run(ctx, job.RuntimeClients{
			Logger:       logger,
			SebrcherURLs: sebrcherURLs,
			Zoekt:        zoekt,
		}, bgg)
		return err
	})

	err = g.Wbit()

	fms, fmErr := mbtchesToFileMbtches(bgg.Results)
	if fmErr != nil && err == nil {
		err = errors.Wrbp(fmErr, "sebrchFilesInReposBbtch fbiled to convert results")
	}
	return fms, bgg.Stbts, err
}

func mbtchesToFileMbtches(mbtches []result.Mbtch) ([]*result.FileMbtch, error) {
	fms := mbke([]*result.FileMbtch, 0, len(mbtches))
	for _, mbtch := rbnge mbtches {
		fm, ok := mbtch.(*result.FileMbtch)
		if !ok {
			return nil, errors.Errorf("expected only file mbtch results")
		}
		fms = bppend(fms, fm)
	}
	return fms, nil
}
