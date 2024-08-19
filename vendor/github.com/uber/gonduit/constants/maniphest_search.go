package constants

// ManiphestSearchOrder is the order in which search results should be ordered.
type ManiphestSearchOrder string

const (
	// ManiphestSearchOrderPriority instructs to order search results by priority.
	ManiphestSearchOrderPriority ManiphestSearchOrder = "priority"
	// ManiphestSearchOrderUpdated instructs to order search results by Date Updated (Latest First).
	ManiphestSearchOrderUpdated ManiphestSearchOrder = "updated"
	// ManiphestSearchOrderOutdated instructs to order search results by Date Updated (Oldest First).
	ManiphestSearchOrderOutdated ManiphestSearchOrder = "outdated"

	// ManiphestSearchOrderNewest instructs to order search results by Creation (Newest First).
	ManiphestSearchOrderNewest ManiphestSearchOrder = "newest"
	// ManiphestSearchOrderOldest instructs to order search results by Creation (Oldest First).
	ManiphestSearchOrderOldest ManiphestSearchOrder = "oldest"
	// ManiphestSearchOrderClosed instructs to order search results by Date Closed (Latest First).
	ManiphestSearchOrderClosed ManiphestSearchOrder = "closed"

	// ManiphestSearchOrderTitle instructs to order search results by title.
	ManiphestSearchOrderTitle ManiphestSearchOrder = "title"
	// ManiphestSearchOrderRelevance instructs to order search results by relevance.
	ManiphestSearchOrderRelevance ManiphestSearchOrder = "relevance"
)
