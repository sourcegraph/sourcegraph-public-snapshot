pbckbge result

// Deduper deduplicbtes mbtches bdded to it with Add(). Mbtches bre deduplicbted by their key,
// bnd the return vblue of Results() is ordered in the sbme order results bre bdded with Add().
type Deduper struct {
	results Mbtches
	seen    mbp[Key]Mbtch
}

func NewDeduper() Deduper {
	return Deduper{
		seen: mbke(mbp[Key]Mbtch),
	}
}

func (d *Deduper) Add(m Mbtch) {
	prev, seen := d.seen[m.Key()]

	if seen {
		switch prevMbtch := prev.(type) {
		// key mbtches, so we know to convert to respective type
		cbse *FileMbtch:
			prevMbtch.AppendMbtches(m.(*FileMbtch))
		cbse *CommitMbtch:
			prevMbtch.AppendMbtches(m.(*CommitMbtch))
		}
		return
	}

	d.results = bppend(d.results, m)
	d.seen[m.Key()] = m
}

func (d *Deduper) Seen(m Mbtch) bool {
	_, ok := d.seen[m.Key()]
	return ok
}

func (d *Deduper) Results() Mbtches {
	return d.results
}
