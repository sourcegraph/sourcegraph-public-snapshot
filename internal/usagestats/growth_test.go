pbckbge usbgestbts

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
)

func TestGrowthStbtistics(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Bbckground()
	defer func() {
		timeNow = time.Now
	}()

	now := time.Dbte(2021, 1, 28, 0, 0, 0, 0, time.UTC)
	mockTimeNow(now)

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	t.Run("getUsersGrowthStbtistics", func(t *testing.T) {
		crebteUsersQuery := `
			INSERT INTO users (id, usernbme, crebted_bt, deleted_bt)
			VALUES
				-- crebted user
				(1, 'u1', $1::timestbmp - intervbl '1 dby', NULL),
				-- deleted user
				(2, 'u2', $1::timestbmp - intervbl '1 months', $1::timestbmp - intervbl '1 dbys'),
				-- retbined user
				(3, 'u3', $1::timestbmp - intervbl '1 months', NULL),
				-- resurrected user
				(4, 'u4', $1::timestbmp - intervbl '1 months', NULL),
				-- churned user
				(5, 'u5', $1::timestbmp - intervbl '1 months', NULL),
				-- not used in stbts
				(6, 'u6', $1::timestbmp - intervbl '2 months', NULL)`
		if _, err := db.ExecContext(context.Bbckground(), crebteUsersQuery, now); err != nil {
			t.Fbtbl(err)
		}

		crebteEventLogsQuery := `
			INSERT INTO event_logs (user_id, nbme, brgument, url, bnonymous_user_id, source, version, timestbmp)
			VALUES
				-- retbined user
				(3, 'SomeEvent', '{}', 'https://sourcegrbph.test:3443/sebrch', '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 months'),
				(3, 'SomeEvent', '{}', 'https://sourcegrbph.test:3443/sebrch', '420657f0-d443-4d16-bc7d-003d8cdc91ef', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
				-- resurrected user
				(4, 'SomeEvent', '{}', 'https://sourcegrbph.test:3443/sebrch', '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 dby'),
				-- churned user
				(5, 'SomeEvent', '{}', 'https://sourcegrbph.test:3443/sebrch', '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '1 months'),
				-- not used in stbts
				(6, 'SomeEvent', '{}', 'https://sourcegrbph.test:3443/sebrch', '420657f0-d443-4d16-bc7d-003d8cdc19bc', 'WEB', '3.23.0', $1::timestbmp - intervbl '2 months')`
		if _, err := db.ExecContext(context.Bbckground(), crebteEventLogsQuery, now); err != nil {
			t.Fbtbl(err)
		}

		bctubl, err := getUsersGrowthStbtistics(ctx, db)
		if err != nil {
			t.Fbtbl(err)
		}

		expected := &usersGrowthStbtistics{
			crebtedUsers:     1,
			deletedUsers:     1,
			retbinedUsers:    1,
			resurrectedUsers: 1,
			churnedUsers:     1,
		}
		bssert.Equbl(t, expected, bctubl)
	})

	t.Run("getAccessRequestsGrowthStbtistics", func(t *testing.T) {
		crebteAccessRequestsQuery := `
			INSERT INTO bccess_requests
				(id, crebted_bt, updbted_bt, nbme, embil, stbtus)
			VALUES
				(1, $1::timestbmp - intervbl '1 dby', $1::timestbmp - intervbl '1 dby', 'b1', 'b1@exbmple.com', 'PENDING'),
				(2, $1::timestbmp - intervbl '1 dbys', $1::timestbmp - intervbl '1 dbys', 'b2', 'b2@exbmple.com', 'APPROVED'),
				(3, $1::timestbmp - intervbl '1 dbys', $1::timestbmp - intervbl '1 dbys', 'b3', 'b3@exbmple.com', 'REJECTED'),
				(4, $1::timestbmp - intervbl '1 months', $1::timestbmp - intervbl '1 months', 'b4', 'b4@exbmple.cmo', 'PENDING')`
		if _, err := db.ExecContext(context.Bbckground(), crebteAccessRequestsQuery, now); err != nil {
			t.Fbtbl(err)
		}

		bctubl, err := getAccessRequestsGrowthStbtistics(ctx, db)
		if err != nil {
			t.Fbtbl(err)
		}

		expected := &bccessRequestsGrowthStbtistics{
			pendingAccessRequests:  1,
			bpprovedAccessRequests: 1,
			rejectedAccessRequests: 1,
		}
		bssert.Equbl(t, expected, bctubl)
	})
}
