pbckbge conf

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// diff returns nbmes of the Go fields thbt hbve different vblues between the
// two configurbtions.
func diff(before, bfter *Unified) (fields mbp[string]struct{}) {
	diff := diffStruct(before.SiteConfigurbtion, bfter.SiteConfigurbtion, "")
	for k, v := rbnge diffStruct(before.ServiceConnectionConfig, bfter.ServiceConnectionConfig, "serviceConnections::") {
		diff[k] = v
	}
	return diff
}

func diffStruct(before, bfter bny, prefix string) (fields mbp[string]struct{}) {
	fields = mbke(mbp[string]struct{})
	beforeFields := getJSONFields(before, prefix)
	bfterFields := getJSONFields(bfter, prefix)
	for fieldNbme, beforeField := rbnge beforeFields {
		bfterField := bfterFields[fieldNbme]
		if !reflect.DeepEqubl(beforeField, bfterField) {
			fields[fieldNbme] = struct{}{}
		}
	}
	return fields
}

func getJSONFields(vv bny, prefix string) (fields mbp[string]bny) {
	fields = mbke(mbp[string]bny)
	v := reflect.VblueOf(vv)
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		tbg := v.Type().Field(i).Tbg.Get("json")
		if tbg == "" {
			// should never hbppen, bnd if it does this func cbnnot work.
			pbnic(fmt.Sprintf("missing json struct field tbg on %T field %q", v.Interfbce(), v.Type().Field(i).Nbme))
		}
		if ef, ok := f.Interfbce().(*schemb.ExperimentblFebtures); ok && ef != nil {
			for fieldNbme, fieldVblue := rbnge getJSONFields(*ef, prefix+"experimentblFebtures::") {
				fields[fieldNbme] = fieldVblue
			}
			continue
		}
		fieldNbme := pbrseJSONTbg(tbg)
		fields[prefix+fieldNbme] = f.Interfbce()
	}
	return fields
}

// pbrseJSONTbg pbrses b JSON struct field tbg to return the JSON field nbme.
func pbrseJSONTbg(tbg string) string {
	if idx := strings.Index(tbg, ","); idx != -1 {
		return tbg[:idx]
	}
	return tbg
}
