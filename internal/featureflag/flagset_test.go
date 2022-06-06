package featureflag

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
)

func TestEvaluatedFlagSet(t *testing.T) {
	cache := map[string][]byte{}
	mockConn := NewMockConn()
	mockConn.DoFunc.SetDefaultHook(func(command string, args ...interface{}) (interface{}, error) {
		switch command {
		case "":
			return nil, nil
		case "HSET":
			cache[args[0].(string)+"_"+args[1].(string)] = []byte(args[2].(string))
		case "HGET":
			return cache[args[0].(string)+"_"+args[1].(string)], nil
		case "DEL":
			delete(cache, args[0].(string)+"_"+args[1].(string))
		default:
			panic("unknown command " + command)
		}
		return nil, nil
	})
	setupRedisTest(t, mockConn)

	e := FlagSet{
		flags: map[string]bool{
			"false1": false,
			"false2": false,
			"false3": false,
			"false4": false,
			"true1":  true,
			"true2":  true,
			"true3":  true,
		},
		actor: actor.FromUser(32),
	}

	got, ok := e.GetBool("true1")
	require.True(t, ok)
	require.True(t, got)

	got, ok = e.GetBool("false1")
	require.True(t, ok)
	require.False(t, got)

	got, ok = e.GetBool("fake")
	require.False(t, ok)

	require.False(t, e.GetBoolOr("false2", true))
	require.False(t, e.GetBoolOr("false3", false))

	require.True(t, e.GetBoolOr("fake", true))
	require.False(t, e.GetBoolOr("fake", false))

	evaluatedFlagSet := getEvaluatedFlagSetFromCache(&e)
	expectedEvaluatedFlagSet := EvaluatedFlagSet{
		"true1":  true,
		"false1": false,
		"false2": false,
		"false3": false,
	}
	require.Equal(t, expectedEvaluatedFlagSet, evaluatedFlagSet)
}
