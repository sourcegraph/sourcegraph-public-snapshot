package resolvers

type ErrIDIsZero struct{}

func (e ErrIDIsZero) Error() string {
	return "invalid node id"
}

func (e ErrIDIsZero) Extensions() map[string]interface{} {
	return map[string]interface{}{"code": "ErrIDIsZero"}
}

type ErrCampaignsDisabled struct{}

func (e ErrCampaignsDisabled) Error() string {
	return "campaigns are disabled. Set 'campaigns.enabled' in the site configuration to enable the feature."
}

func (e ErrCampaignsDisabled) Extensions() map[string]interface{} {
	return map[string]interface{}{"code": "ErrCampaignsDisabled"}
}

type ErrCampaignsDotCom struct{}

func (e ErrCampaignsDotCom) Error() string {
	return "access to campaigns on Sourcegraph.com is currently not available"
}

func (e ErrCampaignsDotCom) Extensions() map[string]interface{} {
	return map[string]interface{}{"code": "ErrCampaignsDotCom"}
}

type ErrEnsureCampaignFailed struct{}

func (e ErrEnsureCampaignFailed) Error() string {
	return "a campaign in the given namespace and with the given name exists but does not match the given ID"
}

func (e ErrEnsureCampaignFailed) Extensions() map[string]interface{} {
	return map[string]interface{}{"code": "ErrEnsureCampaignFailed"}
}

type ErrApplyClosedCampaign struct{}

func (e ErrApplyClosedCampaign) Error() string {
	return "existing campaign matched by campaign spec is closed"
}

func (e ErrApplyClosedCampaign) Extensions() map[string]interface{} {
	return map[string]interface{}{"code": "ErrApplyClosedCampaign"}
}

type ErrMatchingCampaignExists struct{}

func (e ErrMatchingCampaignExists) Error() string {
	return "a campaign matching the given campaign spec already exists"
}

func (e ErrMatchingCampaignExists) Extensions() map[string]interface{} {
	return map[string]interface{}{"code": "ErrMatchingCampaignExists"}
}

type ErrDuplicateCredential struct{}

func (e ErrDuplicateCredential) Error() string {
	return "a credential for this code host already exists"
}

func (e ErrDuplicateCredential) Extensions() map[string]interface{} {
	return map[string]interface{}{"code": "ErrDuplicateCredential"}
}
