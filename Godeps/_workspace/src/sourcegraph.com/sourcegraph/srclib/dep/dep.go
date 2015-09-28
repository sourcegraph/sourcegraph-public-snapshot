package dep

type ResolvedDep struct {
	// FromRepo is the repository from which this dependency originates.
	FromRepo string `json:",omitempty"`

	// FromCommitID is the VCS commit in the repository that this dep was found
	// in.
	FromCommitID string `json:",omitempty"`

	// FromUnit is the source unit name from which this dependency originates.
	FromUnit string 

	// FromUnitType is the source unit type from which this dependency originates.
	FromUnitType string 

	// ToRepo is the repository containing the source unit that is depended on.
	//
	// TODO(sqs): include repo clone URLs as well, so we can add new
	// repositories from seen deps.
	ToRepo string 

	// ToUnit is the name of the source unit that is depended on.
	ToUnit string 

	// ToUnitType is the type of the source unit that is depended on.
	ToUnitType string 

	// ToVersion is the version of the dependent repository (if known),
	// according to whatever version string specifier is used by FromRepo's
	// dependency management system.
	ToVersionString string 

	// ToRevSpec specifies the desired VCS revision of the dependent repository
	// (if known).
	ToRevSpec string 
}
