package mailreply

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/mail"
	"strconv"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
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

		contentType := strings.ToLower(part.Header.Get("Content-Type"))
		if contentType == `text/plain; charset=utf-8` || contentType == `text/plain; charset="utf-8"` {
			slurp, err := ioutil.ReadAll(part)
			if err != nil {
				return nil, errors.Wrap(err, "ReadAll(2)")
			}
			return slurp, nil
		}
	}
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
