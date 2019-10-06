package campaigns

import (
	"encoding/json"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/diagnostics"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
)

type extdata struct{}

func (extdata) parseDiagnostics(d *graphqlbackend.CampaignExtensionData) ([]graphqlbackend.Diagnostic, error) {
	ds := make([]graphqlbackend.Diagnostic, len(d.RawDiagnostics))
	for i, diagnosticStr := range d.RawDiagnostics {
		var d diagnostics.GQLDiagnostic
		if err := json.Unmarshal([]byte(diagnosticStr), &d); err != nil {
			return nil, err
		}
		ds[i] = d
	}
	return ds, nil
}

type diagnosticInfo struct {
	Resource      string
	ResourceURI   gituri.URI
	Message       string
	RawDiagnostic diagnostics.GQLDiagnostic
}

func (extdata) parseDiagnosticInfos(d *graphqlbackend.CampaignExtensionData) ([]diagnosticInfo, error) {
	diags, err := extdata{}.parseDiagnostics(d)
	if err != nil {
		return nil, err
	}

	dis := make([]diagnosticInfo, len(diags))
	for i, diagnostic := range diags {
		if err := json.Unmarshal([]byte(diagnostic.Data().Value.(json.RawMessage)), &dis[i]); err != nil {
			return nil, err
		}
		uri, err := gituri.Parse(dis[i].Resource)
		if err != nil {
			return nil, err
		}
		dis[i].ResourceURI = *uri
		dis[i].RawDiagnostic = diagnostic.(diagnostics.GQLDiagnostic)
	}
	return dis, nil
}

func (extdata) parseRawFileDiffs(d *graphqlbackend.CampaignExtensionData) ([]*diff.FileDiff, error) {
	diffs := make([]*diff.FileDiff, len(d.RawFileDiffs))
	for i, diffStr := range d.RawFileDiffs {
		var err error
		diffs[i], err = diff.ParseFileDiff([]byte(diffStr))
		if err != nil {
			return nil, err
		}
	}
	return diffs, nil
}
