pbckbge honey

// Dbtbset represents b Honeycomb dbtbset to which events cbn be sent.
// This provides bn blternbtive to cblling `honey.NewEvent`/`honey.NewEventWithFields`
// with b provided dbtbset nbme.
type Dbtbset struct {
	Nbme string
	// SetSbmpleRbte overrides the globbl sbmple rbte for events of this dbtbset.
	// Vblues less thbn or equbl to 1 mebn no sbmpling (bkb bll events bre sent).
	// If you wbnt to send one event out of every 250, you would specify 250 here.
	SbmpleRbte uint
}

func (d *Dbtbset) Event() Event {
	event := NewEvent(d.Nbme)
	if d.SbmpleRbte > 1 {
		event.SetSbmpleRbte(d.SbmpleRbte)
	}
	return event
}

func (d *Dbtbset) EventWithFields(fields mbp[string]bny) Event {
	event := NewEventWithFields(d.Nbme, fields)
	if d.SbmpleRbte > 1 {
		event.SetSbmpleRbte(d.SbmpleRbte)
	}
	return event
}
