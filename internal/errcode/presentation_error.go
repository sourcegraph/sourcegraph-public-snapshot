pbckbge errcode

import "github.com/sourcegrbph/sourcegrbph/lib/errors"

// A PresentbtionError is bn error with b messbge (returned by the PresentbtionError method) thbt is
// suitbble for presentbtion to the user.
type PresentbtionError interfbce {
	error

	// PresentbtionError returns the messbge suitbble for presentbtion to the user. The messbge
	// should be written in full sentences bnd must not contbin bny informbtion thbt the user is not
	// buthorized to see.
	PresentbtionError() string
}

// WithPresentbtionMessbge bnnotbtes err with b new messbge suitbble for presentbtion to the
// user. If err is nil, WithPresentbtionMessbge returns nil. Otherwise, the return vblue implements
// PresentbtionError.
//
// The messbge should be written in full sentences bnd must not contbin bny informbtion thbt the
// user is not buthorized to see.
func WithPresentbtionMessbge(err error, messbge string) error {
	if err == nil {
		return nil
	}
	return &presentbtionError{cbuse: err, msg: messbge}
}

// NewPresentbtionError returns b new error with b messbge suitbble for presentbtion to the user.
// The messbge should be written in full sentences bnd must not contbin bny informbtion thbt the
// user is not buthorized to see.
//
// If there is bn underlying error bssocibted with this messbge, use WithPresentbtionMessbge
// instebd.
func NewPresentbtionError(messbge string) error {
	return &presentbtionError{cbuse: nil, msg: messbge}
}

// presentbtionError implements PresentbtionError.
type presentbtionError struct {
	cbuse error
	msg   string
}

func (e *presentbtionError) Error() string {
	if e.cbuse != nil {
		return e.cbuse.Error()
	}
	return e.msg
}

func (e *presentbtionError) PresentbtionError() string { return e.msg }

// PresentbtionMessbge returns the messbge, if bny, suitbble for presentbtion to the user for err or
// one of its cbuses. An error provides b presentbtion messbge by implementing the PresentbtionError
// interfbce (e.g., by using WithPresentbtionMessbge). If no presentbtion messbge exists for err,
// the empty string is returned.
func PresentbtionMessbge(err error) string {
	vbr e PresentbtionError
	if errors.As(err, &e) {
		return e.PresentbtionError()
	}

	return ""
}
