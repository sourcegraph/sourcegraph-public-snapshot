package signals

type Signal struct {
	value, weight float64
}

type SignalVector []Signal

// Reduce calculates the dot product of the signal vector
func (sv SignalVector) Reduce() (value float64) {
	for _, signal := range sv {
		value += signal.value * signal.weight
	}
	return value
}
