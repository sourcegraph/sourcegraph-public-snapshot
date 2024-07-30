package tail

import "strings"

type commandMsg struct {
	name string
	args []string
}

func (c *commandMsg) toPred() activityPred {
	switch c.name {
	case "drop":
		return func(a *activityMsg) *activityMsg {
			switch subject := c.args[0]; subject {
			case "name":
				if strings.HasPrefix(a.name, c.args[1]) {
					return nil
				}
			case "level":
				if strings.EqualFold(a.level, c.args[1]) {
					return nil
				}
			}
			return a
		}
	case "only":
		return func(a *activityMsg) *activityMsg {
			switch subject := c.args[0]; subject {
			case "name":
				if !strings.HasPrefix(a.name, c.args[1]) {
					return nil
				}
			case "level":
				if strings.EqualFold(a.level, c.args[1]) {
					return nil
				}
			}
			return a
		}
	case "grep":
		return func(a *activityMsg) *activityMsg {
			var invert bool
			q := c.args[0]
			if c.args[0] == "-v" {
				invert = true
				q = c.args[1]
			}

			if strings.Contains(a.data, q) != !invert {
				return nil
			}
			return a
		}
	}
	return nil
}
