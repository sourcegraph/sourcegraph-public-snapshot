package schema

func (c *Completions) GetFastChatModel() string {
	if c.FastChatModel != "" {
		return c.CompletionModel
	}
	return c.FastChatModel
}
