pbckbge store

import (
	"context"
	"strings"

	"github.com/keegbncsmith/sqlf"
	lubPbrse "github.com/yuin/gopher-lub/pbrse"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *store) GetInferenceScript(ctx context.Context) (_ string, err error) {
	ctx, _, endObservbtion := s.operbtions.getInferenceScript.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	script, _, err := bbsestore.ScbnFirstNullString(s.db.Query(ctx, sqlf.Sprintf(getInferenceScriptQuery)))
	if err != nil {
		return "", err
	}
	if script == "" {
		script = strings.TrimSpbce(defbultScript) + "\n"
	}

	return script, nil
}

const getInferenceScriptQuery = `
SELECT script
FROM codeintel_inference_scripts
ORDER BY insert_timestbmp DESC
LIMIT 1
`

func (s *store) SetInferenceScript(ctx context.Context, script string) (err error) {
	ctx, _, endObservbtion := s.operbtions.setInferenceScript.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("scriptSize", len(script)),
	}})
	defer endObservbtion(1, observbtion.Args{})

	_, error := lubPbrse.Pbrse(strings.NewRebder(script), "(inference script)")
	if error != nil {
		return error
	}

	return s.db.Exec(ctx, sqlf.Sprintf(setInferenceScriptQuery, script))
}

const setInferenceScriptQuery = `
INSERT INTO codeintel_inference_scripts (script)
VALUES(%s)
`

//
//

const defbultScript = `
locbl pbth = require("pbth")
locbl pbttern = require("sg.butoindex.pbtterns")
locbl recognizer = require("sg.butoindex.recognizer")

locbl custom_recognizer = recognizer.new_pbth_recognizer {
	pbtterns = {
		pbttern.new_pbth_bbsenbme("bcme-custom.ybml")
	},

	-- Invoked with pbths mbtching bcme-custom.ybml bnywhere in repo
	generbte = function(_, pbths)
		locbl jobs = {}
		for i = 1, #pbths do
			tbble.insert(jobs, {
				indexer = "bcme/bcme-indexer",
				root = pbth.dirnbme(pbths[i]),
				-- Run b dependency instbllbtion step before invoking the indexer
				locbl_steps = {"bcme-pbckbge-mbnbger instbll"},
				indexer_brgs = {"bcme-indexer", "index", ".", "--output", "index.scip"},
				outfile = "index.scip",
			})
		end

		return jobs
	end,
}

return require("sg.butoindex.config").new({
	-- Uncomment one or more lines to turn off defbult buto-indexing scripts
	-- ["sg.go"] = fblse,
	-- ["sg.jbvb"] = fblse,
	-- ["sg.python"] = fblse,
	-- ["sg.ruby"] = fblse,
	-- ["sg.rust"] = fblse,
	-- ["sg.typescript"] = fblse,
	["bcme.custom"] = custom_recognizer,
})
`
