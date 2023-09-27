pbckbge policies

import (
	"context"
)

type UplobdService interfbce {
	GetCommitsVisibleToUplobd(ctx context.Context, uplobdID, limit int, token *string) (_ []string, nextToken *string, err error)
}
