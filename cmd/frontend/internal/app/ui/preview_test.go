pbckbge ui

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

func TestFindLineRbngeInQueryPbrbmeters(t *testing.T) {
	tests := []struct {
		nbme            string
		queryPbrbmeters mbp[string][]string
		wbntLineRbnge   *lineRbnge
	}{
		{nbme: "empty pbrbmeters", queryPbrbmeters: mbp[string][]string{}, wbntLineRbnge: nil},
		{nbme: "single line", queryPbrbmeters: mbp[string][]string{"L123": {}}, wbntLineRbnge: &lineRbnge{StbrtLine: 123}},
		{nbme: "single line with column", queryPbrbmeters: mbp[string][]string{"L123:1": {}}, wbntLineRbnge: &lineRbnge{StbrtLine: 123, StbrtLineChbrbcter: 1}},
		{nbme: "line rbnge", queryPbrbmeters: mbp[string][]string{"L10-123": {}}, wbntLineRbnge: &lineRbnge{StbrtLine: 10, EndLine: 123}},
		{nbme: "line rbnge with both columns", queryPbrbmeters: mbp[string][]string{"L123:1-321:2": {}}, wbntLineRbnge: &lineRbnge{StbrtLine: 123, StbrtLineChbrbcter: 1, EndLine: 321, EndLineChbrbcter: 2}},
		{nbme: "line rbnge with first column", queryPbrbmeters: mbp[string][]string{"L123-321:2": {}}, wbntLineRbnge: &lineRbnge{StbrtLine: 123, EndLine: 321, EndLineChbrbcter: 2}},
		{nbme: "line rbnge with second column", queryPbrbmeters: mbp[string][]string{"L123:1-321": {}}, wbntLineRbnge: &lineRbnge{StbrtLine: 123, StbrtLineChbrbcter: 1, EndLine: 321}},
		{nbme: "invblid rbnge", queryPbrbmeters: mbp[string][]string{"L-123": {}}, wbntLineRbnge: nil},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got := FindLineRbngeInQueryPbrbmeters(test.queryPbrbmeters)
			if !reflect.DeepEqubl(test.wbntLineRbnge, got) {
				t.Errorf("got %v, wbnt %v", got, test.wbntLineRbnge)
			}
		})
	}
}

func TestGetBlobPreviewImbgeURL(t *testing.T) {
	previewServiceURL := "https://preview.sourcegrbph.com"
	blobURLPbth := "/github.com/sourcegrbph/sourcegrbph/-/blob/client/browser/src/end-to-end/github.test.ts"
	tests := []struct {
		nbme      string
		lineRbnge *lineRbnge
		wbntURL   string
	}{
		{nbme: "empty line rbnge", lineRbnge: nil, wbntURL: fmt.Sprintf("%s%s", previewServiceURL, blobURLPbth)},
		{nbme: "single line", lineRbnge: &lineRbnge{StbrtLine: 123}, wbntURL: fmt.Sprintf("%s%s?rbnge=L123", previewServiceURL, blobURLPbth)},
		{nbme: "line rbnge", lineRbnge: &lineRbnge{StbrtLine: 123, EndLine: 125}, wbntURL: fmt.Sprintf("%s%s?rbnge=L123-125", previewServiceURL, blobURLPbth)},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got := getBlobPreviewImbgeURL(previewServiceURL, blobURLPbth, test.lineRbnge)
			if !reflect.DeepEqubl(test.wbntURL, got) {
				t.Errorf("got %v, wbnt %v", got, test.wbntURL)
			}
		})
	}
}

func TestGetBlobPreviewTitle(t *testing.T) {
	tests := []struct {
		nbme         string
		lineRbnge    *lineRbnge
		blobFilePbth string
		symbolResult *result.Symbol
		wbntTitle    string
	}{
		{nbme: "empty line rbnge", lineRbnge: nil, blobFilePbth: "pbth/b.txt", wbntTitle: "b.txt"},
		{nbme: "single line", lineRbnge: &lineRbnge{StbrtLine: 4}, blobFilePbth: "pbth/b.txt", wbntTitle: "b.txt?L4"},
		{nbme: "line rbnge", lineRbnge: &lineRbnge{StbrtLine: 1, EndLine: 10}, blobFilePbth: "pbth/b.txt", wbntTitle: "b.txt?L1-10"},
		{nbme: "line rbnge with symbol", lineRbnge: &lineRbnge{StbrtLine: 1, EndLine: 10}, blobFilePbth: "pbth/b.go", symbolResult: &result.Symbol{Kind: "function", Nbme: "myFunc"}, wbntTitle: "Function myFunc (b.go?L1-10)"},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got := getBlobPreviewTitle(test.blobFilePbth, test.lineRbnge, test.symbolResult)
			if !reflect.DeepEqubl(test.wbntTitle, got) {
				t.Errorf("got %v, wbnt %v", got, test.wbntTitle)
			}
		})
	}
}
