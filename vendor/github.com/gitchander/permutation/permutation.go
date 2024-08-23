package permutation

type Interface interface {
	// Len is the number of elements in the collection.
	Len() int
	// Swap swaps the elements with indexes i and j.
	Swap(i, j int)
}

type Permutator struct {
	first bool // It is first state of elements
	v     Interface
	b     []int
}

func New(v Interface) *Permutator {
	n := v.Len()
	if n > 0 {
		n--
	}
	return &Permutator{
		first: true,
		v:     v,
		b:     make([]int, n),
	}
}

func (p *Permutator) Next() bool {

	if p.first {
		p.first = false
		return true
	}

	if n, ok := calcFlipSize(p.b); ok {
		flip(p.v, n) // It is the main flip.
		return true
	}

	// It is the last flip. It helps to return the elements to the begin state.
	flip(p.v, p.v.Len())
	p.first = true
	return false // End of permutations.
}

func calcFlipSize(b []int) (int, bool) {
	for i := range b {
		b[i]++
		if k := i + 2; b[i] < k {
			return k, true
		}
		b[i] = 0
	}
	return 0, false
}

// flip is a function which flips first n elements in the slice (v)
func flip(v Interface, n int) {
	i, j := 0, n-1
	for i < j {
		v.Swap(i, j)
		i, j = i+1, j-1
	}
}
