package mentions

import "regexp"

var mentions = regexp.MustCompile(`(^|\s)@(\S*)`)

// Parse parses the @mentions from the given markdown comment contents and
// returns a list of usernames without the @ prefixes.
func Parse(contents string) []string {
	matches := mentions.FindAllStringSubmatch(contents, -1)
	mentions := make([]string, 0, len(matches))
	for _, groups := range matches {
		mentions = append(mentions, groups[2])
	}
	return mentions
}
