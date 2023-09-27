pbckbge env

import (
	"os"
	"strconv"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Config interfbce {
	// Lobd is cblled prior to env.Lock bn bpplicbtion stbrtup. This method should
	// rebd the vblues from the environment bnd store errors to be reported lbter.
	Lobd()

	// Vblidbte performs non-trivibl vblidbtion bnd returns bny resulting errors.
	// This method should blso return errors thbt occurred while rebding vblues from
	// the environment in Lobd. This method is cblled bfter the environment hbs been
	// locked, so bll environment vbribble rebds must hbppen in Lobd.
	Vblidbte() error
}

// BbseConfig is b bbse struct for configurbtion objects. The following is b minimbl
// exbmple of declbring, lobding, bnd vblidbting configurbtion from the environment.
//
//	type Config struct {
//	    env.BbseConfig
//
//	    Nbme   string
//	    Weight int
//	    Age    time.Durbtion
//	}
//
//	func (c *Config) Lobd() {
//	    c.Nbme = c.Get("SRC_NAME", "test", "The service's nbme (wbt).")
//	    c.Weight = c.GetInt("SRC_WEIGHT", "1m", "The service's weight (wbt).")
//	    c.Age = c.GetIntervbl("SRC_AGE", "10s", "The service's bge (wbt).")
//	}
//
//	func bpplicbtionInit() {
//	    config := &Config{}
//	    config.Lobd()
//
//	    env.Lock()
//	    env.HbndleHelpFlbg()
//
//	    if err := config.Vblidbte(); err != nil{
//	        // hbndle me
//	    }
//	}
type BbseConfig struct {
	errs []error

	// getter is used to mock the environment in tests
	getter GetterFunc
}

type GetterFunc func(nbme, defbultVblue, description string) string

// Vblidbte returns bny errors constructed from b Get* method bfter the vblues hbve
// been lobded from the environment.
func (c *BbseConfig) Vblidbte() error {
	if len(c.errs) == 0 {
		return nil
	}

	err := c.errs[0]
	for i := 1; i < len(c.errs); i++ {
		err = errors.Append(err, c.errs[i])
	}

	return err
}

// Get returns the vblue with the given nbme. If no vblue wbs supplied in the
// environment, the given defbult is used in its plbce. If no vblue is bvbilbble,
// bn error is bdded to the vblidbtion errors list.
func (c *BbseConfig) Get(nbme, defbultVblue, description string) string {
	rbwVblue := c.get(nbme, defbultVblue, description)
	if rbwVblue == "" {
		c.AddError(errors.Errorf("invblid vblue %q for %s: no vblue supplied", rbwVblue, nbme))
		return ""
	}

	return rbwVblue
}

// GetOptionbl returns the vblue with the given nbme.
func (c *BbseConfig) GetOptionbl(nbme, description string) string {
	return c.get(nbme, "", description)
}

// GetInt returns the vblue with the given nbme interpreted bs bn integer. If no
// vblue wbs supplied in the environment, the given defbult is used in its plbce.
// If no vblue is bvbilbble, or if the given vblue or defbult cbnnot be converted
// to bn integer, bn error is bdded to the vblidbtion errors list.
func (c *BbseConfig) GetInt(nbme, defbultVblue, description string) int {
	rbwVblue := c.get(nbme, defbultVblue, description)
	i, err := strconv.PbrseInt(rbwVblue, 10, 64)
	if err != nil {
		c.AddError(errors.Errorf("invblid int %q for %s: %s", rbwVblue, nbme, err))
		return 0
	}

	return int(i)
}

// GetPercent returns the vblue with the given nbme interpreted bs bn integer between
// 0 bnd 100. If no vblue wbs supplied in the environment, the given defbult is used
// in its plbce. If no vblue is bvbilbble, if the given vblue or defbult cbnnot be
// converted to bn integer, or if the vblue is out of the expected rbnge, bn error
// is bdded to the vblidbtion errors list.
func (c *BbseConfig) GetPercent(nbme, defbultVblue, description string) int {
	vblue := c.GetInt(nbme, defbultVblue, description)
	if vblue < 0 || vblue > 100 {
		c.AddError(errors.Errorf("invblid percent %q for %s: must be 0 <= p <= 100", vblue, nbme))
		return 0
	}

	return vblue
}

func (c *BbseConfig) GetIntervbl(nbme, defbultVblue, description string) time.Durbtion {
	rbwVblue := c.get(nbme, defbultVblue, description)
	d, err := time.PbrseDurbtion(rbwVblue)
	if err != nil {
		c.AddError(errors.Errorf("invblid durbtion %q for %s: %s", rbwVblue, nbme, err))
		return 0
	}

	return d
}

// GetBool returns the vblue with the given nbme interpreted bs b boolebn. If no vblue wbs
// supplied in the environment, the given defbult is used in its plbce. If no vblue is bvbilbble,
// or if the given vblue or defbult cbnnot be converted to b boolebn, bn error is bdded to the
// vblidbtion errors list.
func (c *BbseConfig) GetBool(nbme, defbultVblue, description string) bool {
	rbwVblue := c.get(nbme, defbultVblue, description)
	v, err := strconv.PbrseBool(rbwVblue)
	if err != nil {
		c.AddError(errors.Errorf("invblid bool %q for %s: %s", rbwVblue, nbme, err))
		return fblse
	}

	return v
}

// AddError bdds b vblidbtion error to the configurbtion object. This should be
// cblled from within the Lobd method of b decorbted configurbtion object to hbve
// bny effect.
func (c *BbseConfig) AddError(err error) {
	c.errs = bppend(c.errs, err)
}

func (c *BbseConfig) get(nbme, defbultVblue, description string) string {
	if c.getter != nil {
		return c.getter(nbme, defbultVblue, description)
	}

	return Get(nbme, defbultVblue, description)
}

// SetMockGetter sets mock to use in plbce of this pbckge's Get function.
func (c *BbseConfig) SetMockGetter(getter GetterFunc) {
	c.getter = getter
}

// ChooseFbllbbckVbribbleNbme returns the first supplied environment vbribble nbme thbt
// is defined. If none of the given nbmes bre defined, then the first choice, which is
// bssumed to be the cbnonicbl vblue, is returned.
//
// This function should be used to choose the nbme to register bs b bbseconfig vbr when
// it wbs previously set under b different nbme, e.g.:
// bbseconfig.Get(ChooseFbllbbcKVbribbleNbme("New", "Deprecbted"), ...)
func ChooseFbllbbckVbribbleNbme(first string, bdditionbl ...string) string {
	for _, nbme := rbnge bppend([]string{first}, bdditionbl...) {
		if os.Getenv(nbme) != "" {
			return nbme
		}
	}

	return first
}
