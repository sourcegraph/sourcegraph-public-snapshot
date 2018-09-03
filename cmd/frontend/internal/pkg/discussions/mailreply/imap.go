package mailreply

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"golang.org/x/net/html/charset"
)

// Message represents an IMAP message.
type Message struct {
	*imap.Message

	client *client.Client
}

// MarkSeenAndDeleted marks this message as deleted.
func (m *Message) MarkSeenAndDeleted() error {
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(m.SeqNum)
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag, imap.DeletedFlag}
	if err := m.client.Store(seqSet, item, flags, nil); err != nil {
		return errors.Wrap(err, "Store")
	}
	return nil
}

// TextContent returns the text contents of the message body.
func (m *Message) TextContent() ([]byte, error) {
	r := m.GetBody(&imap.BodySectionName{})
	if r == nil {
		return nil, errors.Wrap(errors.New("server didn't return message body"), "GetBody")
	}
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return nil, errors.Wrap(err, "ReadMessage")
	}
	body, err := ioutil.ReadAll(msg.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ReadAll")
	}

	// If we already have a text/plain message body, return it directly.
	if m.BodyStructure.MIMEType == "text" && m.BodyStructure.MIMESubType == "plain" {
		return body, nil
	}

	// If we don't have a multipart message body, we don't know how to find
	// plain text in the message.
	if m.BodyStructure.MIMEType != "multipart" {
		return nil, nil
	}

	boundary, ok := m.BodyStructure.Params["BOUNDARY"]
	if !ok {
		return nil, errors.New("multipart message missing BOUNDARY parameter")
	}
	mr := multipart.NewReader(bytes.NewReader(body), boundary)
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			return nil, nil // couldn't find any plain text
		}
		if err != nil {
			return nil, errors.Wrap(err, "NextPart")
		}

		text, err := messagePartTextContent(part, part.Header)
		if err != nil {
			return nil, errors.Wrap(err, "partTextContent")
		}
		defer part.Close()
		if text != nil {
			slurp, err := ioutil.ReadAll(text)
			if err != nil {
				return nil, errors.Wrap(err, "ReadAll(2)")
			}
			return slurp, nil
		}
	}
}

func messagePartTextContent(part io.ReadCloser, header textproto.MIMEHeader) (io.Reader, error) {
	var (
		mediaType string
		params    map[string]string
		err       error
	)
	if _, ok := header["Content-Type"]; ok {
		mediaType, params, err = mime.ParseMediaType(header.Get("Content-Type"))
		if err != nil {
			return nil, errors.Wrap(err, "ParseMediaType")
		}
	} else {
		mediaType = "text/plain"
		params = map[string]string{"charset": "utf-8"}
	}

	// We are only interested in text content, so eliminate other possibilities
	// now (HTML, etc).
	if mediaType != "text/plain" {
		return nil, fmt.Errorf("content type %q not supported", header.Get("Content-Type"))
	}

	// Handle decoding the content first, if needed.
	var decoded io.Reader
	contentTransferEncoding := header.Get("Content-Transfer-Encoding")
	switch strings.ToLower(contentTransferEncoding) {
	case "", "7bit", "8bit":
		// noop, content is not encoded or is otherwise UTF-8 compatible.
		decoded = part
	case "quoted-printable": // very common "mostly plain text" encoding.
		decoded = quotedprintable.NewReader(part)
	case "base64": // common when e.g. replying from gmail with only emojis
		decoded = base64.NewDecoder(base64.StdEncoding, part)
	default:
		return nil, nil // we do not know how to handle this encoding
	}

	// Handle possible character sets, if needed.
	return charset.NewReaderLabel(strings.ToLower(params["charset"]), decoded)
}

// NewMailReader returns a new reader that reads mail from the configured IMAP
// server. If no IMAP server is configured nil, nil is returned.
func NewMailReader() (*MailReader, error) {
	if !conf.CanReadEmail() {
		return nil, nil
	}
	conf := conf.Get()

	// Connect to the IMAP server.
	c, err := client.DialTLS(net.JoinHostPort(conf.EmailImap.Host, strconv.Itoa(conf.EmailImap.Port)), nil)
	if err != nil {
		return nil, errors.Wrap(err, "DialTLS")
	}

	// Login, if needed.
	if conf.EmailImap.Username != "" {
		if err := c.Login(conf.EmailImap.Username, conf.EmailImap.Password); err != nil {
			return nil, errors.Wrap(err, "Login")
		}
	}

	readOnly := false
	_, err = c.Select("INBOX", readOnly)
	if err != nil {
		return nil, errors.Wrap(err, "Select INBOX")
	}
	return &MailReader{client: c}, nil
}

// MailReader reads mail from an IMAP server.
type MailReader struct {
	client *client.Client
}

// Close closes the reader. No operations can be performed on the reader or
// messages returned from it previously after this method is invoked.
func (r *MailReader) Close() error {
	return r.client.Logout()
}

// ReadUnread reads all unread mail and sends it to the given channel.
func (r *MailReader) ReadUnread(ch chan *Message, done chan error) error {
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	seqNums, err := r.client.Search(criteria)
	if err != nil {
		return errors.Wrap(err, "Search")
	}
	if len(seqNums) == 0 {
		close(ch)
		done <- nil
		close(done)
		return nil
	}
	seqSet := &imap.SeqSet{}
	seqSet.AddNum(seqNums...)

	messages := make(chan *imap.Message, 10)
	go func() {
		section := &imap.BodySectionName{Peek: true}
		done <- r.client.Fetch(seqSet, []imap.FetchItem{imap.FetchEnvelope, imap.FetchBodyStructure, section.FetchItem()}, messages)
		close(done)
	}()
	go func() {
		for msg := range messages {
			ch <- &Message{Message: msg, client: r.client}
		}
		close(ch)
	}()
	return nil
}
