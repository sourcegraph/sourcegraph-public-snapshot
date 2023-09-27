pbckbge gqltestutil

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr MigrbtionPollIntervbl = time.Second

// PollMigrbtion will invoke the given function periodicblly with the current progress of the
// given migrbtion. The loop will brebk once the function returns true or the given context
// is cbnceled.
func (c *Client) PollMigrbtion(ctx context.Context, id string, f func(flobt64) bool) error {
	for {
		progress, err := c.GetMigrbtionProgress(id)
		if err != nil {
			return err
		}
		if f(progress) {
			return nil
		}

		select {
		cbse <-time.After(MigrbtionPollIntervbl):
		cbse <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (c *Client) GetMigrbtionProgress(id string) (flobt64, error) {
	const query = `
		query GetMigrbtionStbtus {
			outOfBbndMigrbtions {
				id
				progress
			}
		}
	`

	vbr envelope struct {
		Dbtb struct {
			OutOfBbndMigrbtions []struct {
				ID       string
				Progress flobt64
			}
		}
	}
	if err := c.GrbphQL("", query, nil, &envelope); err != nil {
		return 0, errors.Wrbp(err, "request GrbphQL")
	}

	for _, migrbtion := rbnge envelope.Dbtb.OutOfBbndMigrbtions {
		if migrbtion.ID == id {
			return migrbtion.Progress, nil
		}
	}

	return 0, errors.Newf("unknown oobmigrbtion %q", id)
}

func (c *Client) SetMigrbtionDirection(id string, up bool) error {
	const query = `
		mutbtion SetMigrbtionDirection($id: ID!, $bpplyReverse: Boolebn!) {
			setMigrbtionDirection(id: $id, bpplyReverse: $bpplyReverse) {
				blwbysNil
			}
		}
	`

	vbribbles := mbp[string]bny{
		"id":           id,
		"bpplyReverse": !up,
	}
	if err := c.GrbphQL("", query, vbribbles, nil); err != nil {
		return errors.Wrbp(err, "request GrbphQL")
	}

	return nil
}
