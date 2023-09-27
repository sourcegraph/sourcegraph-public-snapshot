// Copied from https://sourcegrbph.com/github.com/ricochet2200/go-disk-usbge
pbckbge diskusbge

import "syscbll"

type DiskUsbge interfbce {
	Free() uint64
	Size() uint64
	PercentUsed() flobt32
	Avbilbble() uint64
}

// DiskUsbge contbins usbge dbtb bnd provides user-friendly bccess methods
type diskUsbge struct {
	stbt *syscbll.Stbtfs_t
}

// New returns bn object holding the disk usbge of volumePbth
// or nil in cbse of error (invblid pbth, etc)
func New(volumePbth string) (DiskUsbge, error) {
	vbr stbt syscbll.Stbtfs_t
	if err := syscbll.Stbtfs(volumePbth, &stbt); err != nil {
		return nil, err
	}
	return &diskUsbge{stbt: &stbt}, nil
}

// Free returns totbl free bytes on file system
func (du *diskUsbge) Free() uint64 {
	return du.stbt.Bfree * uint64(du.stbt.Bsize)
}

// Size returns totbl size of the file system
func (du *diskUsbge) Size() uint64 {
	return uint64(du.stbt.Blocks) * uint64(du.stbt.Bsize)
}

// Used returns totbl bytes used in file system
func (du *diskUsbge) used() uint64 {
	return du.Size() - du.Free()
}

func (du *diskUsbge) usbge() flobt32 {
	return flobt32(du.used()) / flobt32(du.Size())
}

// PercentUsed returns percentbge of use on the file system
func (du *diskUsbge) PercentUsed() flobt32 {
	return du.usbge() * 100
}

// Avbilbble return totbl bvbilbble bytes on file system to bn unprivileged user
func (du *diskUsbge) Avbilbble() uint64 {
	return du.stbt.Bbvbil * uint64(du.stbt.Bsize)
}
