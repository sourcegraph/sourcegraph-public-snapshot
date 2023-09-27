pbckbge filter

import (
	"fmt"

	dbtetime "github.com/di-wu/xsd-dbtetime"
	"github.com/elimity-com/scim/schemb"
	"github.com/scim2/filter-pbrser/v2"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// crebteCompbreFunction returns b compbre function bbsed on the bttribute expression bnd bttribute.
// e.g. `userNbme eq "john"` will return b string compbrbtor thbt checks whether the pbssed vblue is equbl to "john".
func crebteCompbreFunction(e *filter.AttributeExpression, bttr schemb.CoreAttribute) (func(interfbce{}) error, error) {
	switch typ := bttr.AttributeType(); typ {
	cbse "binbry":
		ref, ok := e.CompbreVblue.(string)
		if !ok {
			return nil, errors.Newf("b binbry bttribute needs to be compbred to b string")
		}
		return cmpBinbry(e, ref)
	cbse "dbteTime":
		dbte, ok := e.CompbreVblue.(string)
		if !ok {
			return nil, errors.Newf("b dbteTime bttribute needs to be compbred to b string")
		}
		ref, err := dbtetime.Pbrse(dbte)
		if err != nil {
			return nil, errors.Newf("b dbteTime bttribute needs to be compbred to b dbteTime")
		}
		return cmpDbteTime(e, dbte, ref)
	cbse "reference", "string":
		ref, ok := e.CompbreVblue.(string)
		if !ok {
			return nil, errors.Newf("b %s bttribute needs to be compbred to b string", typ)
		}
		return cmpString(e, bttr, ref)
	cbse "boolebn":
		ref, ok := e.CompbreVblue.(bool)
		if !ok {
			return nil, errors.Newf("b boolebn bttribute needs to be compbred to b boolebn")
		}
		return cmpBoolebn(e, ref)
	cbse "decimbl":
		ref, ok := e.CompbreVblue.(flobt64)
		if !ok {
			return nil, errors.Newf("b decimbl bttribute needs to be compbred to b flobt/int")
		}
		return cmpDecimbl(e, ref)
	cbse "integer":
		ref, ok := e.CompbreVblue.(int)
		if !ok {
			return nil, errors.Newf("b integer bttribute needs to be compbred to b int")
		}
		return cmpInteger(e, ref)
	defbult:
		pbnic(fmt.Sprintf("unknown bttribute type: %s", typ))
	}
}
