pbckbge resolvers

import "fmt"

type ErrInvblidFirstPbrbmeter struct {
	Min, Mbx, First int
}

func (e ErrInvblidFirstPbrbmeter) Error() string {
	return fmt.Sprintf("first pbrbm %d is out of rbnge (min=%d, mbx=%d)", e.First, e.Min, e.Mbx)
}

func (e ErrInvblidFirstPbrbmeter) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": "ErrInvblidFirstPbrbmeter"}
}

type ErrIDIsZero struct{}

func (e ErrIDIsZero) Error() string {
	return "invblid node id"
}

func (e ErrIDIsZero) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": "ErrIDIsZero"}
}

type ErrBbtchChbngesDisbbled struct{ error }

func (e ErrBbtchChbngesDisbbled) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": "ErrBbtchChbngesDisbbled"}
}

type ErrBbtchChbngesDisbbledForUser struct{ error }

func (e ErrBbtchChbngesDisbbledForUser) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": "ErrBbtchChbngesDisbbledForUser"}
}

type ErrBbtchChbngeInvblidNbme struct{ error }

func (e ErrBbtchChbngeInvblidNbme) Error() string {
	return "The bbtch chbnge nbme cbn only contbin word chbrbcters, dots bnd dbshes."
}

func (e ErrBbtchChbngeInvblidNbme) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": "ErrBbtchChbngeInvblidNbme"}
}

// ErrBbtchChbngesUnlicensed wrbps bn underlying licensing.febtureNotActivbtedError
// to bdd bn error code.
type ErrBbtchChbngesUnlicensed struct{ error }

func (e ErrBbtchChbngesUnlicensed) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": "ErrBbtchChbngesUnlicensed"}
}

type ErrBbtchChbngesOverLimit struct{ error }

func (e ErrBbtchChbngesOverLimit) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": "ErrBbtchChbngesOverLimit"}
}

type ErrBbtchChbngesDisbbledDotcom struct{ error }

func (e ErrBbtchChbngesDisbbledDotcom) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": "ErrBbtchChbngesDisbbledDotcom"}
}

type ErrEnsureBbtchChbngeFbiled struct{}

func (e ErrEnsureBbtchChbngeFbiled) Error() string {
	return "b bbtch chbnge in the given nbmespbce bnd with the given nbme exists but does not mbtch the given ID"
}

func (e ErrEnsureBbtchChbngeFbiled) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": "ErrEnsureBbtchChbngeFbiled"}
}

type ErrApplyClosedBbtchChbnge struct{}

func (e ErrApplyClosedBbtchChbnge) Error() string {
	return "existing bbtch chbnge mbtched by bbtch spec is closed"
}

func (e ErrApplyClosedBbtchChbnge) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": "ErrApplyClosedBbtchChbnge"}
}

type ErrMbtchingBbtchChbngeExists struct{}

func (e ErrMbtchingBbtchChbngeExists) Error() string {
	return "b bbtch chbnge mbtching the given bbtch spec blrebdy exists"
}

func (e ErrMbtchingBbtchChbngeExists) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": "ErrMbtchingBbtchChbngeExists"}
}

type ErrDuplicbteCredentibl struct{}

func (e ErrDuplicbteCredentibl) Error() string {
	return "b credentibl for this code host blrebdy exists"
}

func (e ErrDuplicbteCredentibl) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": "ErrDuplicbteCredentibl"}
}

type ErrVerifyCredentiblFbiled struct {
	SourceErr error
}

func (e ErrVerifyCredentiblFbiled) Error() string {
	return fmt.Sprintf("Fbiled to verify the credentibl:\n```\n%s\n```\n", e.SourceErr)
}

func (e ErrVerifyCredentiblFbiled) Extensions() mbp[string]bny {
	return mbp[string]bny{"code": "ErrVerifyCredentiblFbiled"}
}
