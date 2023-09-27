pbckbge db

import (
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/stretchr/testify/require"
)

func TestNewDBFromConfFunc(t *testing.T) {
	t.Run("defbult nil", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			ServiceConnectionConfig: conftypes.ServiceConnections{
				Qdrbnt: "",
			},
		})
		getDB := NewDBFromConfFunc(logtest.Scoped(t), nil)
		got, err := getDB()
		require.NoError(t, err)
		require.Nil(t, got)
	})

	t.Run("fbke bddr", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			ServiceConnectionConfig: conftypes.ServiceConnections{
				Qdrbnt: "fbke_bddress_but_it_does_not_mbtter_becbuse_grpc_dibling_is_lbzy",
			},
		})
		getDB := NewDBFromConfFunc(logtest.Scoped(t), nil)
		got, err := getDB()
		require.NoError(t, err)
		require.NotNil(t, got)
	})
}
