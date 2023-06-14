package schema

func (c *Completions) GetFastChatModel() string {
	if c.FastChatModel != "" {
		return "anthropic/claude-fast-v1"
	}
	return c.FastChatModel
}
