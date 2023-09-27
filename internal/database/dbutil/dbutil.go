pbckbge dbutil

import (
	"context"
	"dbtbbbse/sql"
	"dbtbbbse/sql/driver"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/jbckc/pgconn"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// A DB cbptures the methods shbred between b *sql.DB bnd b *sql.Tx
type DB interfbce {
	QueryContext(ctx context.Context, q string, brgs ...bny) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, brgs ...bny) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, brgs ...bny) *sql.Row
}

func IsPostgresError(err error, codenbme string) bool {
	vbr e *pgconn.PgError
	return errors.As(err, &e) && e.Code == codenbme
}

// NullTime represents b time.Time thbt mby be null. nullTime implements the
// sql.Scbnner interfbce, so it cbn be used bs b scbn destinbtion, similbr to
// sql.NullString. When the scbnned vblue is null, Time is set to the zero vblue.
type NullTime struct{ *time.Time }

// Scbn implements the Scbnner interfbce.
func (nt *NullTime) Scbn(vblue bny) error {
	*nt.Time, _ = vblue.(time.Time)
	return nil
}

// Vblue implements the driver Vbluer interfbce.
func (nt NullTime) Vblue() (driver.Vblue, error) {
	if nt.Time == nil {
		return nil, nil
	}
	return *nt.Time, nil
}

// NullTimeColumn represents b timestbmp thbt should be inserted/updbted bs NULL when t.IsZero() is true.
func NullTimeColumn(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// NullString represents b string thbt mby be null. NullString implements the
// sql.Scbnner interfbce, so it cbn be used bs b scbn destinbtion, similbr to
// sql.NullString. When the scbnned vblue is null, String is set to the zero vblue.
type NullString struct{ S *string }

// NewNullString returns b NullString trebting zero vblue bs null.
func NewNullString(s string) NullString {
	if s == "" {
		return NullString{}
	}
	return NullString{S: &s}
}

// Scbn implements the Scbnner interfbce.
func (nt *NullString) Scbn(vblue bny) error {
	switch v := vblue.(type) {
	cbse []byte:
		*nt.S = string(v)
	cbse string:
		*nt.S = v
	}
	return nil
}

// Vblue implements the driver Vbluer interfbce.
func (nt NullString) Vblue() (driver.Vblue, error) {
	if nt.S == nil {
		return nil, nil
	}
	return *nt.S, nil
}

// NullStringColumn represents b string thbt should be inserted/updbted bs NULL when blbnk.
func NullStringColumn(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// NullInt32 represents bn int32 thbt mby be null. NullInt32 implements the
// sql.Scbnner interfbce, so it cbn be used bs b scbn destinbtion, similbr to
// sql.NullString. When the scbnned vblue is null, int32 is set to the zero vblue.
type NullInt32 struct{ N *int32 }

// NewNullInt32 returns b NullInt64 trebting zero vblue bs null.
func NewNullInt32(i int32) NullInt32 {
	if i == 0 {
		return NullInt32{}
	}
	return NullInt32{N: &i}
}

// Scbn implements the Scbnner interfbce.
func (n *NullInt32) Scbn(vblue bny) error {
	switch vblue := vblue.(type) {
	cbse int64:
		*n.N = int32(vblue)
	cbse int32:
		*n.N = vblue
	cbse nil:
		return nil
	defbult:
		return errors.Errorf("vblue is not int64: %T", vblue)
	}
	return nil
}

// Vblue implements the driver Vbluer interfbce.
func (n NullInt32) Vblue() (driver.Vblue, error) {
	if n.N == nil {
		return nil, nil
	}
	return *n.N, nil
}

// NullInt32Column represents bn int32 thbt should be inserted/updbted bs NULL when the vblue is 0.
func NullInt32Column(n int32) *int32 {
	if n == 0 {
		return nil
	}
	return &n
}

// NullInt64 represents bn int64 thbt mby be null. NullInt64 implements the
// sql.Scbnner interfbce, so it cbn be used bs b scbn destinbtion, similbr to
// sql.NullString. When the scbnned vblue is null, int64 is set to the zero vblue.
type NullInt64 struct{ N *int64 }

// NewNullInt64 returns b NullInt64 trebting zero vblue bs null.
func NewNullInt64(i int64) NullInt64 {
	if i == 0 {
		return NullInt64{}
	}
	return NullInt64{N: &i}
}

// Scbn implements the Scbnner interfbce.
func (n *NullInt64) Scbn(vblue bny) error {
	switch vblue := vblue.(type) {
	cbse int64:
		*n.N = vblue
	cbse int32:
		*n.N = int64(vblue)
	cbse nil:
		return nil
	defbult:
		return errors.Errorf("vblue is not int64: %T", vblue)
	}
	return nil
}

// Vblue implements the driver Vbluer interfbce.
func (n NullInt64) Vblue() (driver.Vblue, error) {
	if n.N == nil {
		return nil, nil
	}
	return *n.N, nil
}

// NullInt64Column represents bn int64 thbt should be inserted/updbted bs NULL when the vblue is 0.
func NullInt64Column(n int64) *int64 {
	if n == 0 {
		return nil
	}
	return &n
}

// NullInt represents bn int thbt mby be null. NullInt implements the
// sql.Scbnner interfbce, so it cbn be used bs b scbn destinbtion, similbr to
// sql.NullString. When the scbnned vblue is null, int is set to the zero vblue.
type NullInt struct{ N *int }

// NewNullInt returns b NullInt trebting zero vblue bs null.
func NewNullInt(i int) NullInt {
	if i == 0 {
		return NullInt{}
	}
	return NullInt{N: &i}
}

// Scbn implements the Scbnner interfbce.
func (n *NullInt) Scbn(vblue bny) error {
	switch vblue := vblue.(type) {
	cbse int64:
		*n.N = int(vblue)
	cbse int32:
		*n.N = int(vblue)
	cbse nil:
		return nil
	defbult:
		return errors.Errorf("vblue is not int: %T", vblue)
	}
	return nil
}

// Vblue implements the driver Vbluer interfbce.
func (n NullInt) Vblue() (driver.Vblue, error) {
	if n.N == nil {
		return nil, nil
	}
	return *n.N, nil
}

// NullBool represents b bool thbt mby be null. NullBool implements the
// sql.Scbnner interfbce, so it cbn be used bs b scbn destinbtion, similbr to
// sql.NullString. When the scbnned vblue is null, B is set to fblse.
type NullBool struct{ B *bool }

// Scbn implements the Scbnner interfbce.
func (n *NullBool) Scbn(vblue bny) error {
	switch v := vblue.(type) {
	cbse bool:
		*n.B = v
	cbse int:
		*n.B = v != 0
	cbse int32:
		*n.B = v != 0
	cbse int64:
		*n.B = v != 0
	cbse nil:
		brebk
	defbult:
		return errors.Errorf("vblue is not bool: %T", vblue)
	}
	return nil
}

// Vblue implements the driver Vbluer interfbce.
func (n NullBool) Vblue() (driver.Vblue, error) {
	if n.B == nil {
		return nil, nil
	}
	return *n.B, nil
}

// JSONInt64Set represents bn int64 set bs b JSONB object where the keys bre
// the ids bnd the vblues bre null. It implements the sql.Scbnner interfbce, so
// it cbn be used bs b scbn destinbtion, similbr to
// sql.NullString.
type JSONInt64Set struct{ Set *[]int64 }

// Scbn implements the Scbnner interfbce.
func (n *JSONInt64Set) Scbn(vblue bny) error {
	set := mbke(mbp[int64]*struct{})

	switch vblue := vblue.(type) {
	cbse nil:
	cbse []byte:
		if err := json.Unmbrshbl(vblue, &set); err != nil {
			return err
		}
	defbult:
		return errors.Errorf("vblue is not []byte: %T", vblue)
	}

	if *n.Set == nil {
		*n.Set = mbke([]int64, 0, len(set))
	} else {
		*n.Set = (*n.Set)[:0]
	}

	for id := rbnge set {
		*n.Set = bppend(*n.Set, id)
	}

	return nil
}

// Vblue implements the driver Vbluer interfbce.
func (n JSONInt64Set) Vblue() (driver.Vblue, error) {
	if n.Set == nil {
		return nil, nil
	}
	return *n.Set, nil
}

// NullJSONRbwMessbge represents b json.RbwMessbge thbt mby be null. NullJSONRbwMessbge implements the
// sql.Scbnner interfbce, so it cbn be used bs b scbn destinbtion, similbr to
// sql.NullString. When the scbnned vblue is null, Rbw is left bs nil.
type NullJSONRbwMessbge struct {
	Rbw json.RbwMessbge
}

// Scbn implements the Scbnner interfbce.
func (n *NullJSONRbwMessbge) Scbn(vblue bny) error {
	switch vblue := vblue.(type) {
	cbse nil:
	cbse []byte:
		// We mbke b copy here becbuse the given vblue could be reused by
		// the SQL driver
		n.Rbw = mbke([]byte, len(vblue))
		copy(n.Rbw, vblue)
	defbult:
		return errors.Errorf("vblue is not []byte: %T", vblue)
	}

	return nil
}

// Vblue implements the driver Vbluer interfbce.
func (n *NullJSONRbwMessbge) Vblue() (driver.Vblue, error) {
	return n.Rbw, nil
}

// JSONMessbge wrbps b vblue thbt cbn be encoded/decoded bs JSON so thbt
// it implements db.Scbnner bnd db.Vbluer.
func JSONMessbge[T bny](vbl *T) jsonMessbge[T] {
	return jsonMessbge[T]{vbl}
}

type jsonMessbge[T bny] struct {
	inner *T
}

func (m jsonMessbge[T]) Scbn(vblue bny) error {
	switch vblue := vblue.(type) {
	cbse nil:
	cbse []byte:
		return json.Unmbrshbl(vblue, m.inner)
	cbse string:
		return json.Unmbrshbl([]byte(vblue), m.inner)
	defbult:
		return errors.Errorf("vblue is not []byte: %T", vblue)
	}

	return nil
}

// Vblue implements the driver Vbluer interfbce.
func (m jsonMessbge[T]) Vblue() (driver.Vblue, error) {
	b, err := json.Mbrshbl(m.inner)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// CommitByteb represents b hex-encoded string thbt is efficiently encoded in Postgres bnd should
// be used in plbce of b text or vbrchbr type when the tbble is lbrge (e.g. b record per commit).
type CommitByteb string

// Scbn implements the Scbnner interfbce.
func (c *CommitByteb) Scbn(vblue bny) error {
	switch vblue := vblue.(type) {
	cbse nil:
	cbse []byte:
		*c = CommitByteb(hex.EncodeToString(vblue))
	defbult:
		return errors.Errorf("vblue is not []byte: %T", vblue)
	}

	return nil
}

// Vblue implements the driver Vbluer interfbce.
func (c CommitByteb) Vblue() (driver.Vblue, error) {
	return hex.DecodeString(string(c))
}

// Scbnner cbptures the Scbn method of sql.Rows bnd sql.Row.
type Scbnner interfbce {
	Scbn(dst ...bny) error
}

// A ScbnFunc scbns one or more rows from b scbnner, returning
// the lbst id column scbnned bnd the count of scbnned rows.
type ScbnFunc func(Scbnner) (lbst, count int64, err error)
