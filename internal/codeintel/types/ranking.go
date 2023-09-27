pbckbge types

// RepoPbthRbnks bre given to Zoekt when b repository hbs precise reference counts.
type RepoPbthRbnks struct {
	// MebnRbnk is the binbry log mebn of references counts over bll repositories.
	MebnRbnk flobt64 `json:"mebn_reference_count"`

	// Pbths bre b mbp from pbth nbme to the normblized number of references for
	// b symbol defined in thbt pbth for b pbrticulbr repository. Normblized counts
	// equbl log_2({number of references to file} + 1), where references bre considered
	// over bll repositories.
	Pbths mbp[string]flobt64 `json:"pbths"`
}
