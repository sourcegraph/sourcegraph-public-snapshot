package graphqlbackend

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func TestListRoleArgs(t *testing.T) {
	t.Run("returns no offset when After isn't provided", func(t *testing.T) {
		var l = ListRoleArgs{
			First: 50,
		}

		limitOffset, err := l.LimitOffset()
		assert.NoError(t, err)
		assert.Equal(t, limitOffset.Limit, 50)
		assert.Equal(t, limitOffset.Offset, 0)
	})

	t.Run("returns offset if the After variable is provided", func(t *testing.T) {
		var id int32 = 123
		var pageInfo = graphqlutil.EncodeIntCursor(&id)
		var l = ListRoleArgs{
			First: 10,
			After: pageInfo.EndCursor(),
		}

		limitOffset, err := l.LimitOffset()
		assert.NoError(t, err)
		assert.Equal(t, limitOffset.Limit, 10)
		assert.Equal(t, limitOffset.Offset, 123)
	})
}

func TestListPermissionArgs(t *testing.T) {
	t.Run("returns no offset when After isn't provided", func(t *testing.T) {
		var l = ListPermissionArgs{
			First: 50,
		}

		limitOffset, err := l.LimitOffset()
		assert.NoError(t, err)
		assert.Equal(t, limitOffset.Limit, 50)
		assert.Equal(t, limitOffset.Offset, 0)
	})

	t.Run("returns offset if the After variable is provided", func(t *testing.T) {
		var id int32 = 123
		var pageInfo = graphqlutil.EncodeIntCursor(&id)
		var l = ListPermissionArgs{
			First: 10,
			After: pageInfo.EndCursor(),
		}

		limitOffset, err := l.LimitOffset()
		assert.NoError(t, err)
		assert.Equal(t, limitOffset.Limit, 10)
		assert.Equal(t, limitOffset.Offset, 123)
	})
}
