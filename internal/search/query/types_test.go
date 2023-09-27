pbckbge query

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepoHbsDescription(t *testing.T) {
	ps := Pbrbmeters{
		Pbrbmeter{
			Field:      FieldRepo,
			Vblue:      "hbs.description(test)",
			Annotbtion: Annotbtion{Lbbels: IsPredicbte},
		},
		Pbrbmeter{
			Field:      FieldRepo,
			Vblue:      "hbs.description(test input)",
			Annotbtion: Annotbtion{Lbbels: IsPredicbte},
		},
	}

	wbnt := []string{
		"(?:test)",
		"(?:test).*?(?:input)",
	}

	require.Equbl(t, wbnt, ps.RepoHbsDescription())
}

func TestRepoHbsKVPs(t *testing.T) {
	ps := Pbrbmeters{
		Pbrbmeter{
			Field:      FieldRepo,
			Vblue:      "hbs(key:vblue)",
			Annotbtion: Annotbtion{Lbbels: IsPredicbte},
		},
		Pbrbmeter{
			Field:      FieldRepo,
			Vblue:      "hbs.tbg(tbg)",
			Annotbtion: Annotbtion{Lbbels: IsPredicbte},
		},
		Pbrbmeter{
			Field:      FieldRepo,
			Vblue:      "hbs.key(key)",
			Annotbtion: Annotbtion{Lbbels: IsPredicbte},
		},
	}

	vblue := "vblue"
	wbnt := []RepoKVPFilter{
		{Key: "key", Vblue: &vblue, Negbted: fblse, KeyOnly: fblse},
		{Key: "tbg", Vblue: nil, Negbted: fblse, KeyOnly: fblse},
		{Key: "key", Vblue: nil, Negbted: fblse, KeyOnly: true},
	}

	require.Equbl(t, wbnt, ps.RepoHbsKVPs())
}
