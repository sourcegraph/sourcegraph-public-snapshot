// Package imap implements IMAP4rev1 (RFC 3501).
package imap

import (
	"errors"
	"io"
	"strings"
)

// A StatusItem is a mailbox status data item that can be retrieved with a
// STATUS command. See RFC 3501 section 6.3.10.
type StatusItem string

const (
	StatusMessages    StatusItem = "MESSAGES"
	StatusRecent                 = "RECENT"
	StatusUidNext                = "UIDNEXT"
	StatusUidValidity            = "UIDVALIDITY"
	StatusUnseen                 = "UNSEEN"
)

// A FetchItem is a message data item that can be fetched.
type FetchItem string

// List of items that can be fetched.
const (
	// Macros
	FetchAll  FetchItem = "ALL"
	FetchFast           = "FAST"
	FetchFull           = "FULL"

	// Items
	FetchBody          = "BODY"
	FetchBodyStructure = "BODYSTRUCTURE"
	FetchEnvelope      = "ENVELOPE"
	FetchFlags         = "FLAGS"
	FetchInternalDate  = "INTERNALDATE"
	FetchRFC822        = "RFC822"
	FetchRFC822Header  = "RFC822.HEADER"
	FetchRFC822Size    = "RFC822.SIZE"
	FetchRFC822Text    = "RFC822.TEXT"
	FetchUid           = "UID"
)

// Expand expands the item if it's a macro.
func (item FetchItem) Expand() []FetchItem {
	switch item {
	case FetchAll:
		return []FetchItem{FetchFlags, FetchInternalDate, FetchRFC822Size, FetchEnvelope}
	case FetchFast:
		return []FetchItem{FetchFlags, FetchInternalDate, FetchRFC822Size}
	case FetchFull:
		return []FetchItem{FetchFlags, FetchInternalDate, FetchRFC822Size, FetchEnvelope, FetchBody}
	default:
		return []FetchItem{item}
	}
}

// FlagsOp is an operation that will be applied on message flags.
type FlagsOp string

const (
	// SetFlags replaces existing flags by new ones.
	SetFlags FlagsOp = "FLAGS"
	// AddFlags adds new flags.
	AddFlags = "+FLAGS"
	// RemoveFlags removes existing flags.
	RemoveFlags = "-FLAGS"
)

// silentOp can be appended to a FlagsOp to prevent the operation from
// triggering unilateral message updates.
const silentOp = ".SILENT"

// A StoreItem is a message data item that can be updated.
type StoreItem string

// FormatFlagsOp returns the StoreItem that executes the flags operation op.
func FormatFlagsOp(op FlagsOp, silent bool) StoreItem {
	s := string(op)
	if silent {
		s += silentOp
	}
	return StoreItem(s)
}

// ParseFlagsOp parses a flags operation from StoreItem.
func ParseFlagsOp(item StoreItem) (op FlagsOp, silent bool, err error) {
	itemStr := string(item)
	silent = strings.HasSuffix(itemStr, silentOp)
	if silent {
		itemStr = strings.TrimSuffix(itemStr, silentOp)
	}
	op = FlagsOp(itemStr)

	if op != SetFlags && op != AddFlags && op != RemoveFlags {
		err = errors.New("Unsupported STORE operation")
	}
	return
}

// CharsetReader, if non-nil, defines a function to generate charset-conversion
// readers, converting from the provided charset into UTF-8. Charsets are always
// lower-case. utf-8 and us-ascii charsets are handled by default. One of the
// the CharsetReader's result values must be non-nil.
var CharsetReader func(charset string, r io.Reader) (io.Reader, error)
