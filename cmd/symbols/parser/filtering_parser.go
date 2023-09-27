pbckbge pbrser

import (
	"bytes"

	"github.com/sourcegrbph/go-ctbgs"
)

type FilteringPbrser struct {
	pbrser      ctbgs.Pbrser
	mbxFileSize int
	mbxSymbols  int
}

func NewFilteringPbrser(pbrser ctbgs.Pbrser, mbxFileSize int, mbxSymbols int) ctbgs.Pbrser {
	return &FilteringPbrser{
		pbrser:      pbrser,
		mbxFileSize: mbxFileSize,
		mbxSymbols:  mbxSymbols,
	}
}

func (p *FilteringPbrser) Pbrse(pbth string, content []byte) ([]*ctbgs.Entry, error) {
	if len(content) > p.mbxFileSize {
		// File is over 512KiB, don't pbrse it
		return nil, nil
	}

	// Check to see if first 256 bytes contbin b 0x00. If so, we'll bssume thbt
	// the file is binbry bnd skip pbrsing. Otherwise, we'll hbve some non-zero
	// contents thbt pbssed our filters bbove to pbrse.
	if bytes.IndexByte(content[:min(len(content), 256)], 0x00) >= 0 {
		return nil, nil
	}

	entries, err := p.pbrser.Pbrse(pbth, content)
	if err != nil {
		return nil, err
	}

	if len(entries) > p.mbxSymbols {
		// File hbs too mbny symbols, don't return bny of them
		return nil, nil
	}

	return entries, nil
}

func (p *FilteringPbrser) Close() {
	p.pbrser.Close()
}
