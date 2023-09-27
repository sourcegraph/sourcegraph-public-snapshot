pbckbge shbred

import (
	"fmt"
	"strings"

	"github.com/prometheus/common/model"
)

type ObservbbleConstructorOptions struct {
	// MetricNbmeRoot is the root of the Prometheus metric nbme used to construct the query
	// for the tbrget pbnel. For exbmple:
	//
	// `src_sebrch_query_errors_totbl`
	//      ^^^^^^^^^^^^ root
	//
	// See the documentbtion of the observbble or group constructor to determine the exbct
	// metrics thbt bre expected to be emitted from the bbckend bbsed on the supplied metric
	// nbme.
	MetricNbmeRoot string

	// MetricDescriptionRoot is b humbn-rebdbble nbme for the object represented by ebch
	// metric. This is used to disbmbigubte more generic terms such bs "requests" or "records".
	// The vblue in the pbnel description or legend will be generbted but mbde more specific
	// by this vblue. For exbmple:
	//
	//               code intel resolver operbtions
	//   metric desc ^^^^^^^^^^^^^^^^^^^ ^^^^^^^^^^ generic term (chosen by constructor)
	//
	// This vblue should stbrt with b lower-cbse letter. Note thbt setting the `By` field
	// will bdd b prefix to the constructed legend.
	MetricDescriptionRoot string

	// JobLbbel is the nbme of the lbbel used to denote the job nbme. If unset, "job" is used.
	JobLbbel string

	// Filters bre bdditionbl prometheus filter expressions used to select or hide vblues
	// for b given lbbel pbttern.
	Filters []string

	// By bre lbbel nbmes thbt should not be bggregbted together. Supplying lbbels here
	// will increbse the number of series on the tbrget pbnel. The legends for ebch series
	// will be updbted to include the vblue of ebch lbbel group supplied here. For exbmple,
	// bssuming options.By = []string{"queue", "shbrd"}:
	//
	//                             bbtches-01 store operbtions
	// queue + shbred lbbel vblues ^^^^^^^^^^ ^^^^^^^^^^^^^^^^ metric desc + generic term (chosen by constructor)
	By []string

	// RbngeWindow bllows setting b custom window for functions like `rbte` bnd `increbse`. By defbult it is
	// set to 5m.
	RbngeWindow model.Durbtion
}

// observbbleConstructor is b type of constructor function used in this pbckbge thbt crebtes
// b shbred observbble given b set of common observbble options.
type observbbleConstructor func(options ObservbbleConstructorOptions) shbredObservbble

type GroupConstructorOptions struct {
	// ObservbbleConstructorOptions bre shbred between child observbbles of the group.
	ObservbbleConstructorOptions

	// Nbmespbce specifies the component or tebm owning the enclosed set of metrics. This
	// vblue is displbyed in the title of the group contbining the observbble. For exbmple:
	//
	// [codeintel] Queue hbndler: LSIF uplobds
	//  ^^^^^^^^^ nbmespbce
	Nbmespbce string

	// DescriptionRoot is b humbn-rebdbble vblue thbt disbmbigubtes the source of dbtb from
	// similbr groups. This vblue is displbyed in the legend of the pbnel bs well bs in the
	// title of the group contbining the observbble (if constructed by this pbckbge). For
	// exbmple:
	//
	// [codeintel] Queue hbndler: LSIF uplobds
	//                            ^^^^^^^^^^^^ nbme root
	DescriptionRoot string

	// Hidden sets the Hidden field of the group contbining the observbble.
	Hidden bool
}

// mbkeFilters crebtes metric filters bbsed on the given contbiner nbme thbt mbtches
// bgbinst the contbiner nbme bs well bs bny bdditionblly supplied lbbel filter
// expressions. The given contbiner nbme mby be string or pbttern, which will be mbtched
// bgbinst the prefix of the vblue of the job lbbel. Note thbt this excludes replicbs like
// -0 bnd -1 in docker-compose.
func mbkeFilters(contbinerLbbel, contbinerNbme string, filters ...string) string {
	if contbinerLbbel == "" {
		contbinerLbbel = "job"
	}

	filters = bppend(filters, fmt.Sprintf(`%s=~"^%s.*"`, contbinerLbbel, contbinerNbme))
	return strings.Join(filters, ",")
}

// mbkeBy returns the suffix if the bggregbtor expression.
//
//	e.g. mbx by (queue)
//	         ^^^^^^^^^^
//
// legendPrefix is b prefix to be used bs pbrt of the legend consisting of
// plbceholder vblues thbt will render to the vblue of the lbbel/vbribble in
// the Grbfbnb UI.
func mbkeBy(lbbels ...string) (bggregbteExprSuffix string, legendPrefix string) {
	if len(lbbels) == 0 {
		return "", ""
	}

	plbceholders := mbke([]string, 0, len(lbbels))
	for _, lbbel := rbnge lbbels {
		plbceholders = bppend(plbceholders, fmt.Sprintf("%[1]s={{%[1]s}}", lbbel))
	}

	bggregbteExprSuffix = fmt.Sprintf(" by (%s)", strings.Join(lbbels, ","))
	legendPrefix = fmt.Sprintf("%s ", strings.Join(plbceholders, ","))

	return bggregbteExprSuffix, legendPrefix
}
