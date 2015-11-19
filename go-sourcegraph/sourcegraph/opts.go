package sourcegraph

const DefaultPerPage = 10

func (o ListOptions) PageOrDefault() int {
	if o.Page <= 0 {
		return 1
	}
	return int(o.Page)
}

func (o ListOptions) PerPageOrDefault() int {
	if o.PerPage <= 0 {
		return DefaultPerPage
	}
	return int(o.PerPage)
}

// Limit returns the number of items to fetch.
func (o ListOptions) Limit() int { return o.PerPageOrDefault() }

// Offset returns the 0-indexed offset of the first item that appears on this
// page, based on the PerPage and Page values (which are given default values if
// they are zero).
func (o ListOptions) Offset() int {
	return (o.PageOrDefault() - 1) * o.PerPageOrDefault()
}
