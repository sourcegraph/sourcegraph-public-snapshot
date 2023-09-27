//go:build !go1.10
// +build !go1.10

pbckbge sebrch_test

import "brchive/tbr"

func bddpbxhebder(w *tbr.Writer, body string) error {
	hdr := &tbr.Hebder{
		Nbme:     "pbx_globbl_hebder",
		Size:     int64(len(body)),
		Typeflbg: tbr.TypeXGlobblHebder,
	}
	if err := w.WriteHebder(hdr); err != nil {
		return err
	}
	_, err := w.Write([]byte(body))
	return err
}
