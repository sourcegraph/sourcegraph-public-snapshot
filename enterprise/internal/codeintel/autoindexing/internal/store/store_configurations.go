package store

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// GetIndexConfigurationByRepositoryID returns the index configuration for a repository.
func (s *store) GetIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int) (_ shared.IndexConfiguration, _ bool, err error) {
	ctx, _, endObservation := s.operations.getIndexConfigurationByRepositoryID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	return scanFirstIndexConfiguration(s.db.Query(ctx, sqlf.Sprintf(getIndexConfigurationByRepositoryIDQuery, repositoryID)))
}

const getIndexConfigurationByRepositoryIDQuery = `
SELECT
	c.id,
	c.repository_id,
	c.data
FROM lsif_index_configuration c WHERE c.repository_id = %s
`

// UpdateIndexConfigurationByRepositoryID updates the index configuration for a repository.
func (s *store) UpdateIndexConfigurationByRepositoryID(ctx context.Context, repositoryID int, data []byte) (err error) {
	ctx, _, endObservation := s.operations.updateIndexConfigurationByRepositoryID.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
	}})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(updateIndexConfigurationByRepositoryIDQuery, repositoryID, data, data))
}

const updateIndexConfigurationByRepositoryIDQuery = `
INSERT INTO lsif_index_configuration (repository_id, data) VALUES (%s, %s)
	ON CONFLICT (repository_id) DO UPDATE SET data = %s
`

func (s *store) SetInferenceScript(ctx context.Context, script string) (err error) {
	ctx, _, endObservation := s.operations.setInferenceScript.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(setInferenceScriptQuery, script))
}

const setInferenceScriptQuery = `
INSERT INTO codeintel_inference_scripts(script)
VALUES(%s)
`

func (s *store) GetInferenceScript(ctx context.Context) (_ string, err error) {
	ctx, _, endObservation := s.operations.getInferenceScript.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	script, _, err := basestore.ScanFirstNullString(s.db.Query(ctx, sqlf.Sprintf(getInferenceScriptQuery)))
	if err != nil {
		return "", err
	}

	if script == "" {
		script = strings.TrimSpace(defaultScript) + "\n"
	}

	return script, nil
}

const getInferenceScriptQuery = `
SELECT script FROM codeintel_inference_scripts
ORDER BY insert_timestamp DESC
LIMIT 1
`

const defaultScript = `
local path = require("path")
local pattern = require("sg.autoindex.patterns")
local recognizer = require("sg.autoindex.recognizer")

local custom_recognizer = recognizer.new_path_recognizer {
	patterns = {
		pattern.new_path_basename("acme-custom.yaml")
	},

	-- Invoked with paths matching acme-custom.yaml anywhere in repo
	generate = function(_, paths)
		local jobs = {}
		for i = 1, #paths do
			table.insert(jobs, {
				steps = {},
				root = path.dirname(paths[i]),
				indexer = "acme/custom-indexer",
				indexer_args = {},
				outfile = "",
			})
		end

		return jobs
	end,
}

return require("sg.autoindex.config").new({
	-- ["sg.clang"] = false,
	-- ["sg.go"] = false,
	-- ["sg.java"] = false,
	-- ["sg.python"] = false,
	-- ["sg.ruby"] = false,
	-- ["sg.rust"] = false,
	-- ["sg.typescript"] = false,
	["acme.custom"] = custom_recognizer,
})
`
