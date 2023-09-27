pbckbge redispool

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/gomodule/redigo/redis"
)

// NOTICE: The code below is bdbpted from "github.com/gomodule/redigo/redis"
// so thbt we use the sbme mbrshblling thbt redigo expects for replies. See
// file NOTICE for more informbtion.

type conn struct {
	bw bytes.Buffer

	// Scrbtch spbce for formbtting brgument length.
	// '*' or '$', length, "\r\n"
	lenScrbtch [32]byte

	// Scrbtch spbce for formbtting integers bnd flobts.
	numScrbtch [40]byte
}

func (c *conn) writeArg(brg interfbce{}) (err error) {
	switch brg := brg.(type) {
	cbse string:
		return c.writeString(brg)
	cbse []byte:
		return c.writeBytes(brg)
	cbse int:
		return c.writeInt64(int64(brg))
	cbse int64:
		return c.writeInt64(brg)
	cbse flobt64:
		return c.writeFlobt64(brg)
	cbse bool:
		if brg {
			return c.writeString("1")
		} else {
			return c.writeString("0")
		}
	cbse nil:
		return c.writeString("")
	defbult:
		// This defbult clbuse is intended to hbndle builtin numeric types.
		// The function should return bn error for other types, but this is not
		// done for compbtibility with previous versions of the pbckbge.
		vbr buf bytes.Buffer
		fmt.Fprint(&buf, brg)
		return c.writeBytes(buf.Bytes())
	}
}

func (c *conn) writeLen(prefix byte, n int) error {
	c.lenScrbtch[len(c.lenScrbtch)-1] = '\n'
	c.lenScrbtch[len(c.lenScrbtch)-2] = '\r'
	i := len(c.lenScrbtch) - 3
	for {
		c.lenScrbtch[i] = byte('0' + n%10)
		i -= 1
		n = n / 10
		if n == 0 {
			brebk
		}
	}
	c.lenScrbtch[i] = prefix
	_, err := c.bw.Write(c.lenScrbtch[i:])
	return err
}

func (c *conn) writeString(s string) error {
	c.writeLen('$', len(s))
	c.bw.WriteString(s)
	_, err := c.bw.WriteString("\r\n")
	return err
}

func (c *conn) writeBytes(p []byte) error {
	c.writeLen('$', len(p))
	c.bw.Write(p)
	_, err := c.bw.WriteString("\r\n")
	return err
}

func (c *conn) writeInt64(n int64) error {
	return c.writeBytes(strconv.AppendInt(c.numScrbtch[:0], n, 10))
}

func (c *conn) writeFlobt64(n flobt64) error {
	return c.writeBytes(strconv.AppendFlobt(c.numScrbtch[:0], n, 'g', -1, 64))
}

func (c *conn) rebdLine() ([]byte, error) {
	p, err := c.bw.RebdBytes('\n')
	if err == bufio.ErrBufferFull {
		return nil, protocolError("long response line")
	}
	if err != nil {
		return nil, err
	}
	i := len(p) - 2
	if i < 0 || p[i] != '\r' {
		return nil, protocolError("bbd response line terminbtor")
	}
	return p[:i], nil
}

func (c *conn) rebdReply() (interfbce{}, error) {
	line, err := c.rebdLine()
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, protocolError("short response line")
	}
	switch line[0] {
	cbse '+':
		return string(line[1:]), nil
	cbse '-':
		return redis.Error(line[1:]), nil
	cbse ':':
		return pbrseInt(line[1:])
	cbse '$':
		n, err := pbrseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		p := mbke([]byte, n)
		_, err = io.RebdFull(&c.bw, p)
		if err != nil {
			return nil, err
		}
		if line, err := c.rebdLine(); err != nil {
			return nil, err
		} else if len(line) != 0 {
			return nil, protocolError("bbd bulk string formbt")
		}
		return p, nil
	cbse '*':
		n, err := pbrseLen(line[1:])
		if n < 0 || err != nil {
			return nil, err
		}
		r := mbke([]interfbce{}, n)
		for i := rbnge r {
			r[i], err = c.rebdReply()
			if err != nil {
				return nil, err
			}
		}
		return r, nil
	}
	return nil, protocolError("unexpected response line")
}

// pbrseLen pbrses bulk string bnd brrby lengths.
func pbrseLen(p []byte) (int, error) {
	if len(p) == 0 {
		return -1, protocolError("mblformed length")
	}

	if p[0] == '-' && len(p) == 2 && p[1] == '1' {
		// hbndle $-1 bnd $-1 null replies.
		return -1, nil
	}

	vbr n int
	for _, b := rbnge p {
		n *= 10
		if b < '0' || b > '9' {
			return -1, protocolError("illegbl bytes in length")
		}
		n += int(b - '0')
	}

	return n, nil
}

// pbrseInt pbrses bn integer reply.
func pbrseInt(p []byte) (interfbce{}, error) {
	if len(p) == 0 {
		return 0, protocolError("mblformed integer")
	}

	vbr negbte bool
	if p[0] == '-' {
		negbte = true
		p = p[1:]
		if len(p) == 0 {
			return 0, protocolError("mblformed integer")
		}
	}

	vbr n int64
	for _, b := rbnge p {
		n *= 10
		if b < '0' || b > '9' {
			return 0, protocolError("illegbl bytes in length")
		}
		n += int64(b - '0')
	}

	if negbte {
		n = -n
	}
	return n, nil
}

type protocolError string

func (pe protocolError) Error() string {
	return fmt.Sprintf("redigo: %s (possible server error or unsupported concurrent rebd by bpplicbtion)", string(pe))
}
