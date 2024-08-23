package valast

type cycleDetector struct {
	seen map[interface{}]int
}

func (c *cycleDetector) push(ptr interface{}) bool {
	if c.seen == nil {
		c.seen = map[interface{}]int{}
	}
	cycles, seen := c.seen[ptr]
	if seen && cycles > 1 {
		return true
	}
	c.seen[ptr] = cycles + 1
	return false
}

func (c *cycleDetector) pop(ptr interface{}) {
	cycles := c.seen[ptr]
	cycles--
	if cycles < 0 {
		cycles = 0
	}
	c.seen[ptr] = cycles
}
