package graphqlbackend

type repoVisibility string

const (
	Any     repoVisibility = "any"
	Private repoVisibility = "private"
	Public  repoVisibility = "public"
)

func parseVisibility(s string) repoVisibility {
	switch s {
	case "private":
		return Private
	case "public":
		return Public
	default:
		return Any
	}
}
