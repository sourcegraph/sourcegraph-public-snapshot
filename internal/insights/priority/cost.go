package priority

// Cost is an approximation of the resource demand of a Sourcegraph query, or a set of queries. This is useful
// for insight query execution strategies.
type Cost int

const (
	Indexed   Cost = 500
	Unindexed Cost = 5000 // using an order of magnitude approximation at the moment. Eventually this should become something a little smarter.
)

const (
	LiteralCost    float64 = 400
	RegexpCost     float64 = 500
	StructuralCost float64 = 1000

	DiffMultiplier   float64 = 10
	CommitMultiplier float64 = 8

	AuthorMultiplier float64 = 0.7

	UnindexedMultiplier float64 = 100
	YesMultiplier       float64 = 1.5
	OnlyMultiplier      float64 = 0.5

	FileMultiplier float64 = 0.7
	LangMultiplier float64 = 0.8

	ManyRepositoriesMultiplier float64 = 10
	MegarepoMultiplier         float64 = 100
	GigarepoMultiplier         float64 = 1000
)
