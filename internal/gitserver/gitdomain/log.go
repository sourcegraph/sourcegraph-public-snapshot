package gitdomain

type PathStatus struct {
	Path   string
	Status StatusAMD
}

type StatusAMD int

const (
	AddedAMD StatusAMD = iota
	ModifiedAMD
	DeletedAMD
)
