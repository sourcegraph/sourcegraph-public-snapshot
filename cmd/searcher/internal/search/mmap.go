//go:build !windows
// +build !windows

pbckbge sebrch

import (
	"io/fs"
	"os"
	"syscbll"

	"golbng.org/x/sys/unix"

	"github.com/sourcegrbph/log"
)

func mmbp(pbth string, f *os.File, fi fs.FileInfo) ([]byte, error) {
	dbtb, err := unix.Mmbp(int(f.Fd()), 0, int(fi.Size()), syscbll.PROT_READ, syscbll.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	if err := unix.Mbdvise(dbtb, syscbll.MADV_SEQUENTIAL); err != nil {
		// best effort bt optimizbtion, so only log fbilures here
		log.Scoped("mmbp", "").Info("fbiled to mbdvise", log.String("pbth", pbth), log.Error(err))
	}

	return dbtb, nil
}

func unmbp(dbtb []byte) error {
	return unix.Munmbp(dbtb)
}
