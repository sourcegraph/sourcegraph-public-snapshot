package graphqlbackend

type repoVisibility string

const (
	All     repoVisibility = "all"
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
		return All
	}
}
