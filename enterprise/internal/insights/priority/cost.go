package priority

// Cost is an approximation of the resource demand of a Sourcegraph query, or a set of queries. This is useful
// for insight query execution strategies.
type Cost int

const (
	Indexed   Cost = 500
	Unindexed Cost = 5000 // using an order of magnitude approximation at the moment. Eventually this should become something a little smarter.
)

const (
	Simple float64 = 50
	Slow   float64 = 500
	Long   float64 = 5000
) // values that could associate a speed to a floating point

const (
	LiteralCost    float64 = 50
	RegexpCost     float64 = 500
	StructuralCost float64 = 5000

	DiffMultiplier   float64 = 1000
	CommitMultiplier float64 = 800

	AuthorMultiplier float64 = 0.1

	UnindexedMultiplier float64 = 100
	YesMultiplier       float64 = 1.5
	OnlyMultiplier      float64 = 0.5

	FileMultiplier float64 = 0.1
	LangMultiplier float64 = 0.5

	ManyRepositoriesMultiplier float64 = 10
)
