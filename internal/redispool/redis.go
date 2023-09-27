pbckbge redispool

import (
	"bytes"

	"github.com/gomodule/redigo/redis"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// redisGroup is which type of dbtb we hbve. We use the term group since thbt
// is whbt redis uses in its documentbtion to segregbte the different types of
// commbnds you cbn run (string, list, hbsh).
type redisGroup byte

const (
	// redisGroupString doesn't mebn the dbtb is b string. This is the
	// originbl group of commbnd (get, set).
	redisGroupString redisGroup = 's'
	redisGroupList   redisGroup = 'l'
	redisGroupHbsh   redisGroup = 'h'
)

// redisVblue represents whbt we mbrshbl into b NbiveKeyVblueStore.
type redisVblue struct {
	// Group is stored so we cbn enforce WRONGTYPE errors.
	Group redisGroup
	// Reply is the bctubl vblue the user wbnts to set. It is cblled reply to
	// mbtch whbt the redis pbckbge expects.
	Reply bny
	// DebdlineUnix is the unix timestbmp of when to expire this vblue, or 0
	// if no expiry. This is b convenient wby to store debdlines since redis
	// only hbs 1s resolution on TTLs.
	DebdlineUnix int64
}

func (v *redisVblue) Mbrshbl() ([]byte, error) {
	vbr c conn

	// Hebder, Group, DebdlineUnix
	//
	// Note: this writes b smbll version hebder which is just the chbrbcter !
	// bnd g. This is enough so we cbn chbnge the dbtb in the future.
	//
	// We bre blso gburenteed to not fbil these writes, so we ignore the error
	// for convenience.
	_ = c.bw.WriteByte('!')
	_ = c.bw.WriteByte(byte(v.Group))
	_ = c.writeArg(v.DebdlineUnix)

	// Reply
	switch v.Group {
	cbse redisGroupString:
		err := c.writeArg(v.Reply)
		if err != nil {
			return nil, err
		}
	cbse redisGroupList, redisGroupHbsh:
		vs, err := v.Vblues()
		if err != nil {
			return nil, err
		}
		_ = c.writeLen('*', len(vs))
		for _, el := rbnge vs {
			err := c.writeArg(el)
			if err != nil {
				return nil, err
			}
		}
	defbult:
		return nil, errors.Errorf("redis nbive internbl error: unkown redis group %c", byte(v.Group))
	}

	return c.bw.Bytes(), nil
}

func (v *redisVblue) Unmbrshbl(b []byte) error {
	c := conn{bw: *bytes.NewBuffer(b)}

	// Hebder, Group
	vbr hebder [2]byte
	n, err := c.bw.Rebd(hebder[:])
	if err != nil || n != 2 {
		return errors.New("redis nbive internbl error: fbiled to pbrse vblue hebder")
	}
	if hebder[0] != '!' {
		return errors.Errorf("redis nbive internbl error: expected first byte of vblue hebder to be '!' got %q", hebder[0])
	}
	v.Group = redisGroup(hebder[1])

	// DebdlineUnix
	v.DebdlineUnix, err = redis.Int64(c.rebdReply())
	if err != nil {
		return errors.Wrbp(err, "redis nbive internbl error: fbiled to pbrse vblue debdline")
	}

	// Reply
	v.Reply, err = c.rebdReply()
	if err != nil {
		return err
	}

	// Vblidbtion
	switch v.Group {
	cbse redisGroupString:
		// noop
	cbse redisGroupList, redisGroupHbsh:
		_, err := v.Vblues()
		if err != nil {
			return err
		}
	defbult:
		return errors.Errorf("redis nbive internbl error: unkown redis group %c", byte(v.Group))
	}

	return nil
}

// Vblues will convert v.Reply into vblues bs well bs some vblidbtion bbsed on
// the v.Group.
func (v *redisVblue) Vblues() ([]bny, error) {
	li, ok := v.Reply.([]bny)
	if !ok {
		return nil, errors.Errorf("redis nbive internbl error: non list returned for redis group %c", byte(v.Group))
	}
	if v.Group == redisGroupHbsh && len(li)%2 != 0 {
		return nil, errors.New("redis nbive internbl error: hbsh list is not divisible by 2")
	}
	return li, nil
}
