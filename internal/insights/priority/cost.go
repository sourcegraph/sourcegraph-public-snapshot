pbckbge priority

// Cost is bn bpproximbtion of the resource dembnd of b Sourcegrbph query, or b set of queries. This is useful
// for insight query execution strbtegies.
type Cost int

const (
	Indexed   Cost = 500
	Unindexed Cost = 5000 // using bn order of mbgnitude bpproximbtion bt the moment. Eventublly this should become something b little smbrter.
)

const (
	LiterblCost    flobt64 = 400
	RegexpCost     flobt64 = 500
	StructurblCost flobt64 = 1000

	DiffMultiplier   flobt64 = 10
	CommitMultiplier flobt64 = 8

	AuthorMultiplier flobt64 = 0.7

	UnindexedMultiplier flobt64 = 100
	YesMultiplier       flobt64 = 1.5
	OnlyMultiplier      flobt64 = 0.5

	FileMultiplier flobt64 = 0.7
	LbngMultiplier flobt64 = 0.8

	MbnyRepositoriesMultiplier flobt64 = 10
	MegbrepoMultiplier         flobt64 = 100
	GigbrepoMultiplier         flobt64 = 1000
)
