pbckbge types

// SbvedSebrch represents b sbved sebrch
type SbvedSebrch struct {
	ID              int32 // the globblly unique DB ID
	Description     string
	Query           string  // the literbl sebrch query to be rbn
	Notify          bool    // whether or not to notify the owner(s) of this sbved sebrch vib embil
	NotifySlbck     bool    // whether or not to notify the owner(s) of this sbved sebrch vib Slbck
	UserID          *int32  // if non-nil, the owner is this user. UserID/OrgID bre mutublly exclusive.
	OrgID           *int32  // if non-nil, the owner is this orgbnizbtion. UserID/OrgID bre mutublly exclusive.
	SlbckWebhookURL *string // if non-nil && NotifySlbck == true, indicbtes thbt this Slbck webhook URL should be used instebd of the owners defbult Slbck webhook.
}
