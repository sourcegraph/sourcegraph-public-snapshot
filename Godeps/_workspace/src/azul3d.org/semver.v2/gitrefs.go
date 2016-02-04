// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package semver

import (
	"bytes"
	"fmt"
	"strings"
)

// Note: A unofficial specification for the Git Smart/Dumb HTTP protocols is
// available at:
//
//  https://gist.github.com/schacon/6092633
//

type gitRef struct {
	Name, Hash string

	// Only present if this is a peeled ref ("^{}").
	PeeledHash string
}

// BestHash returns the PeeledHash, if present, otherwise the Hash.
func (r *gitRef) BestHash() string {
	if len(r.PeeledHash) > 0 {
		return r.PeeledHash
	}
	return r.Hash
}

type gitRefs struct {
	service  string
	mainID   string // e.g. SHA1 of HEAD
	mainName string // e.g. "HEAD"
	capList  []string
	records  []*gitRef
}

func (r *gitRefs) Bytes() []byte {
	var b []byte

	// The service line and it's line-break (0000).
	b = append(b, gitPktLine(fmt.Sprintf("# service=%s\n", r.service)).Bytes()...)
	b = append(b, []byte("0000")...)

	// The mainID + mainName + capList.
	mmc := gitPktLine(r.mainID)
	mmc = append(mmc, ' ')
	mmc = append(mmc, []byte(r.mainName)...)
	mmc = append(mmc, '\x00')
	if len(r.capList) > 0 {
		mmc = append(mmc, []byte(strings.Join(r.capList, " "))...)
	}
	mmc = append(mmc, '\n')
	b = append(b, mmc.Bytes()...)

	// The ref records.
	for _, ref := range r.records {
		// First the unpeeled one.
		unpeeled := gitPktLine(ref.Hash)
		unpeeled = append(unpeeled, ' ')
		unpeeled = append(unpeeled, []byte(ref.Name)...)
		unpeeled = append(unpeeled, '\n')
		b = append(b, unpeeled.Bytes()...)

		// If it exists, also add the peeled one.
		if len(ref.PeeledHash) > 0 {
			peeled := gitPktLine(ref.PeeledHash)
			peeled = append(peeled, ' ')
			peeled = append(peeled, []byte(ref.Name)...)
			peeled = append(peeled, []byte("^{}\n")...)
			b = append(b, peeled.Bytes()...)
		}
	}

	// Close the records with a line-breal.
	b = append(b, []byte("0000")...)
	return b
}

func gitParseRefs(infoRefs []byte) (*gitRefs, error) {
	parser := new(gitRefParser)
	parser.refs = new(gitRefs)
	parser.next = parser.parseServiceSection

	for {
		// Decode the next pkt-line.
		pl, lineBreak, n, err := gitNextPktLine(infoRefs)
		if err != nil {
			return nil, err
		}
		infoRefs = infoRefs[n:]

		// Feed to parser.
		err = parser.next(pl, lineBreak)
		if err != nil {
			return nil, err
		}
		if parser.done {
			break
		}
	}
	return parser.refs, nil
}

type gitRefParser struct {
	refs                 *gitRefs
	done                 bool
	next, afterLineBreak func(pl gitPktLine, lineBreak bool) error
	lastRefRecord        *[2]string
}

// expectLineBreak expects a single line break ("0000") pkt-line and then
// proceeds to p.afterLineBreak upon success.
func (p *gitRefParser) expectLineBreak(pl gitPktLine, lineBreak bool) error {
	if !lineBreak {
		return fmt.Errorf("expected line break")
	}
	p.next = p.afterLineBreak
	p.afterLineBreak = nil
	return nil
}

// parseServiceSection parses the initial service pkt-line of the smart_reply.
func (p *gitRefParser) parseServiceSection(pl gitPktLine, lineBreak bool) error {
	// Trim space of the line.
	pl = bytes.TrimSpace(pl)

	// Split the bytes based on "service=" and just grab the right side.
	split := bytes.Split(pl, []byte("service="))
	if len(split) == 0 {
		return fmt.Errorf("parseServiceSection: expected service pkt-line")
	}
	p.refs.service = string(split[1])

	// Next up we expect a single line break and then the ref_list.
	p.next = p.expectLineBreak
	p.afterLineBreak = p.parseRefList
	return nil
}

// parseRefList parses the ref_list portion of the smart_reply.
func (p *gitRefParser) parseRefList(pl gitPktLine, lineBreak bool) error {
	// Trim space of the line.
	pl = bytes.TrimSpace(pl)

	// At this point we have:
	// "227b26555939499162b40a7ab64265e70cd3a790 HEAD\x00multi_ack_detailed multi_ack side-band-64k thin-pack ofs-delta no-progress include-tag shallow\n"

	// Cut out the ID SHA1 hash.
	if len(pl) < 41 {
		return fmt.Errorf("parseRefList: expected prefixed main ID SHA1 hash")
	}
	p.refs.mainID = string(pl[:40])
	pl = pl[41:] // 40 + 1 space

	// At this point we have:
	// "HEAD\x00multi_ack_detailed multi_ack side-band-64k thin-pack ofs-delta no-progress include-tag shallow\n"

	// Split based on the NULL character to grab the left-side name.
	nulSplit := bytes.Split(pl, []byte{'\x00'})
	if len(nulSplit) < 2 {
		return fmt.Errorf("parseRefList: expected NUL seperation byte")
	}
	p.refs.mainName = string(nulSplit[0])
	pl = pl[len(p.refs.mainName)+1:]

	// At this point we have:
	// "multi_ack_detailed multi_ack side-band-64k thin-pack ofs-delta no-progress include-tag shallow\n"

	// Split the cap_list based on spaces, strip the suffixed newline.
	pl = bytes.TrimSuffix(pl, []byte{'\n'})
	capList := bytes.Split(pl, []byte{' '})
	for _, c := range capList {
		p.refs.capList = append(p.refs.capList, string(c))
	}

	// Check for empty lists.
	if p.refs.mainName == "capabilities{}^" {
		// empty list: we are done parsing.
		// Note: be aware that a "0000" line-break does still follow.
		p.done = true
		return nil
	}

	// No empty list? Ref records follow then:
	p.next = p.parseRefRecords
	return nil
}

// parseRefRecords parses the ref_records portion of the smart_reply.
func (p *gitRefParser) parseRefRecords(pl gitPktLine, lineBreak bool) error {
	if lineBreak {
		// If there is a last record, we can insert it into the records slice
		// now as an unpeeled ref.
		if p.lastRefRecord != nil {
			p.refs.records = append(p.refs.records, &gitRef{
				Hash: p.lastRefRecord[0],
				Name: p.lastRefRecord[1],
			})
		}
		p.done = true
		return nil
	}

	// The line looks like: "cd95fa968a0fa851547bd65e73e1b385a2dca005 refs/heads/master\n"

	// Trim the suffixed newline.
	pl = bytes.TrimSuffix(pl, []byte{'\n'})

	// Split by the space in the middle.
	split := bytes.Split(pl, []byte{' '})
	if len(split) != 2 {
		return fmt.Errorf("parseRefRecords: expected space seperated value")
	}
	s := [2]string{
		string(split[0]), // e.g. "cd95fa968a0fa851547bd65e73e1b385a2dca005"
		string(split[1]), // e.g. "refs/heads/master"
	}

	// If there is no last record, then store this one for later and return
	// immedietly.
	if p.lastRefRecord == nil {
		p.lastRefRecord = &s
		return nil
	}

	// We have a last record at this point, which means we can derive it's type
	// (peeled or unpeeled).
	name := s[1]
	if strings.HasSuffix(name, "^{}") {
		// We have a peeled reference.
		p.refs.records = append(p.refs.records, &gitRef{
			Hash:       p.lastRefRecord[0],
			Name:       p.lastRefRecord[1],
			PeeledHash: s[0],
		})
		p.lastRefRecord = nil
		return nil
	}

	// We have a unpeeled reference.
	p.refs.records = append(p.refs.records, &gitRef{
		Hash: p.lastRefRecord[0],
		Name: p.lastRefRecord[1],
	})
	p.lastRefRecord = &s
	return nil
}
