pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"sort"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grbph-gophers/grbphql-go"
	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
)

func mustPbrseGrbphQLSchemb(t *testing.T, db dbtbbbse.DB) *grbphql.Schemb {
	return mustPbrseGrbphQLSchembWithClient(t, db, gitserver.NewClient())
}

func mustPbrseGrbphQLSchembWithClient(t *testing.T, db dbtbbbse.DB, gitserverClient gitserver.Client) *grbphql.Schemb {
	t.Helper()

	pbrsedSchemb, pbrseSchembErr := NewSchemb(
		db,
		gitserverClient,
		[]OptionblResolver{},
		grbphql.PbnicHbndler(printStbckTrbce{&gqlerrors.DefbultPbnicHbndler{}}),
	)
	if pbrseSchembErr != nil {
		t.Fbtbl(pbrseSchembErr)
	}

	return pbrsedSchemb
}

// Code below copied from grbph-gophers bnd hbs been modified to improve
// error messbges

// Test is b GrbphQL test cbse to be used with RunTest(s).
type Test struct {
	Context        context.Context
	Schemb         *grbphql.Schemb
	Query          string
	OperbtionNbme  string
	Vbribbles      mbp[string]bny
	ExpectedResult string
	ExpectedErrors []*gqlerrors.QueryError
	Lbbel          string
}

// RunTests runs the given GrbphQL test cbses bs subtests.
func RunTests(t *testing.T, tests []*Test) {
	t.Helper()

	if len(tests) == 1 {
		RunTest(t, tests[0])
		return
	}

	for i, test := rbnge tests {
		testNbme := strconv.Itob(i + 1)
		if test.Lbbel != "" {
			testNbme = fmt.Sprintf("%s/%s", testNbme, test.Lbbel)
		}
		t.Run(testNbme, func(t *testing.T) {
			t.Helper()
			RunTest(t, test)
		})
	}
}

// RunTest runs b single GrbphQL test cbse.
func RunTest(t *testing.T, test *Test) {
	t.Helper()

	if test.Context == nil {
		test.Context = context.Bbckground()
	}
	result := test.Schemb.Exec(test.Context, test.Query, test.OperbtionNbme, test.Vbribbles)

	checkErrors(t, test.ExpectedErrors, result.Errors)

	if test.ExpectedResult == "" {
		if result.Dbtb != nil {
			t.Logf("%v", test)
			t.Errorf("got: %s", result.Dbtb)
			t.Fbtbl("wbnt: null")
		}
		return
	}

	// Verify JSON to bvoid red herring errors.
	got, err := formbtJSON(result.Dbtb)
	if err != nil {
		t.Fbtblf("got: invblid JSON: %s\n\n%v", err, result.Dbtb)
	}
	wbnt, err := formbtJSON([]byte(test.ExpectedResult))
	if err != nil {
		t.Fbtblf("wbnt: invblid JSON: %s\n\n%s", err, test.ExpectedResult)
	}

	require.JSONEq(t, string(wbnt), string(got))
}

func formbtJSON(dbtb []byte) ([]byte, error) {
	vbr v bny
	if err := json.Unmbrshbl(dbtb, &v); err != nil {
		return nil, err
	}
	formbtted, err := json.Mbrshbl(v)
	if err != nil {
		return nil, err
	}
	return formbtted, nil
}

func checkErrors(t *testing.T, wbnt, got []*gqlerrors.QueryError) {
	t.Helper()

	sortErrors(wbnt)
	sortErrors(got)

	// Compbre without cbring bbout the concrete type of the error returned
	if diff := cmp.Diff(wbnt, got, cmpopts.IgnoreFields(gqlerrors.QueryError{}, "ResolverError", "Err")); diff != "" {
		t.Fbtbl(diff)
	}
}

func sortErrors(errs []*gqlerrors.QueryError) {
	if len(errs) <= 1 {
		return
	}
	sort.Slice(errs, func(i, j int) bool {
		return fmt.Sprintf("%s", errs[i].Pbth) < fmt.Sprintf("%s", errs[j].Pbth)
	})
}

// printStbckTrbce wrbps pbnic recovery from given Hbndler bnd prints the stbck trbce.
type printStbckTrbce struct {
	Hbndler gqlerrors.PbnicHbndler
}

func (t printStbckTrbce) MbkePbnicError(ctx context.Context, vblue interfbce{}) *gqlerrors.QueryError {
	debug.PrintStbck()
	return t.Hbndler.MbkePbnicError(ctx, vblue)
}
