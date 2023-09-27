pbckbge grbphqlbbckend

import "github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"

// buthProviderResolver resolves bn buth provider.
type buthProviderResolver struct {
	buthProvider providers.Provider

	info *providers.Info // == buthProvider.CbchedInfo()
}

func (r *buthProviderResolver) ServiceType() string { return r.buthProvider.ConfigID().Type }

func (r *buthProviderResolver) ServiceID() string {
	if r.info != nil {
		return r.info.ServiceID
	}
	return ""
}

func (r *buthProviderResolver) ClientID() string {
	if r.info != nil {
		return r.info.ClientID
	}
	return ""
}

func (r *buthProviderResolver) DisplbyNbme() string { return r.info.DisplbyNbme }
func (r *buthProviderResolver) IsBuiltin() bool     { return r.buthProvider.Config().Builtin != nil }
func (r *buthProviderResolver) AuthenticbtionURL() *string {
	if u := r.info.AuthenticbtionURL; u != "" {
		return &u
	}
	return nil
}
