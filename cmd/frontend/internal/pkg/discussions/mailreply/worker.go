package mailreply

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/discussions"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// StartWorker should be invoked only after the DB has been initialized. It
// starts the background worker which is responsible for reading mail and
// updating discussion threads based on email replies.
//
// It should be invoked in a separate goroutine.
func StartWorker() {
	conf.Watch(func() {
		if !conf.CanReadEmail() {
			return
		}

		// Only one frontend instance should ever run this worker, so we use a
		// distributed lock to guarantee this. If the frontend with the lock
		// acquired dies, it will be released after 1 minute.
		for {
			var (
				release func()
				ok      bool
				ctx     context.Context
			)
			ctx, release, ok = rcache.TryAcquireMutex(context.Background(), "discussionsMailReplyWorker")
			if !ok {
				// Failed to acquire the mutex. Wait before trying again.
				time.Sleep(30 * time.Second)
				continue
			}

			// Acquired the mutex, perform work under it.
			log15.Debug("discussions: mailreply worker running")
			workForever(ctx)
			log15.Debug("discussions: mailreply worker stopped", "ctx", ctx.Err())
			release()
		}
	})
}

func workForever(ctx context.Context) {
	work := func() error {
		reader, err := NewMailReader()
		if err != nil {
			return errors.Wrap(err, "NewMailReader")
		}
		defer reader.Close()
		done := make(chan error, 1)
		ch := make(chan *Message, 10)
		if err := reader.ReadUnread(ch, done); err != nil {
			return errors.Wrap(err, "ReadUnread")
		}
		for msg := range ch {
			// 🚨 SECURITY: Check that one of the messages "to" addresses
			// includes a valid sub-address authorization token. e.g.
			// "notifications+SomeSecret123@sourcegraph.com". This guarantees
			// that this email came from the user we sent the notification to
			// previously (whereas e.g. relying on the "From" address field
			// would be completely insecure doing to being easily spoofed).
			//
			// See https://tools.ietf.org/html/rfc5233 for details on sub-addressing.
			var (
				haveAuthorization bool
				userID            int32
				threadID          int64
			)
			for _, toAddress := range msg.Envelope.To {
				// Parse the token ("SomeSecret123") out of the mailbox name ("notifications+SomeSecret123").
				split := strings.Split(toAddress.MailboxName, "+")
				if len(split) < 2 {
					continue
				}
				token := split[len(split)-1]

				// Verify the token.
				userID, threadID, err = db.DiscussionMailReplyTokens.Get(ctx, token)
				if err == db.ErrInvalidToken {
					log15.Debug("discussions: mailreply worker: ignoring email with invalid authorization token", "subject", msg.Envelope.Subject, "mailbox_name", toAddress.MailboxName)
					msg.MarkSeenAndDeleted()
					break // Invalid token / attacker
				}
				if err != nil {
					log15.Error("discussions: mailreply worker: error while looking up token", "error", err)
					continue
				}
				haveAuthorization = true
				break
			}
			if !haveAuthorization {
				continue // ignore the message
			}

			textContent, err := msg.TextContent()
			if err != nil {
				log15.Error("discussions: mailreply worker: error while reading TextContent", "error", err)
				continue
			}

			contents := strings.TrimSpace(string(trimGmailReplyQuote(textContent)))
			if contents == "" {
				log15.Debug("discussions: mailreply worker: ignoring email with no effective content", "subject", msg.Envelope.Subject, "content", string(textContent))
				msg.MarkSeenAndDeleted()
				continue // ignore empty replies
			}

			_, err = discussions.InsecureAddCommentToThread(ctx, &types.DiscussionComment{
				ThreadID:     threadID,
				AuthorUserID: userID,
				Contents:     contents,
			})
			if err != nil {
				log15.Error("discussions: mailreply worker: error while adding comment to thread", "error", err)
				continue
			}

			// Now that we're finished handling this message, mark it as seen
			// and to be deleted.
			msg.MarkSeenAndDeleted()
		}
		if err := <-done; err != nil {
			return errors.Wrap(err, "done")
		}
		return nil
	}
	for {
		if ctx.Err() != nil {
			return // e.g. if we lost the distributed mutex
		}
		if err := work(); err != nil {
			log15.Error("discussions: mailreply worker: error while working", "error", err)
		}
		time.Sleep(5 * time.Second)
	}
}

var gmailQuoteMatch = regexp.MustCompile(`(\r\n|\n).*On .* at .*, (.|\r\n|\n)*wrote\:(.|\r\n|\n)*(\r\n|\n)+(>.*(\r\n|\n))+(.|\r\n|\n)*`)

// trimGmailReplyQuote trims the gmail reply quotation out of the given
// message. This is a best-effort approach. In specific, it looks for the
// pattern:
//
// 	On $DATE at $TIME, $NAME $EMAIL wrote:
// 	> $ANYTHING
//  > $ANYTHING
//
// When found, everything past that line is removed.
func trimGmailReplyQuote(m []byte) []byte {
	matches := gmailQuoteMatch.FindAllIndex(m, -1)
	if len(matches) == 0 {
		return m
	}
	firstMatch := matches[0]
	return m[:firstMatch[0]]
}
