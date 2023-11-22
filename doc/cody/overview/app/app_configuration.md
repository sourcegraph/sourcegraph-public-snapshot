# Cody App Configuration

## Using a third-party LLM provider

Instead of using Sourcegraph Cody Gateway, you can configure the Cody app to use a third-party provide directly. Currently this can only be Anthropic or OpenAI.

You must create your own key with Anthropic [here](https://console.anthropic.com/account/keys) or with OpenAI [here](https://beta.openai.com/account/api-keys). Once you have the key, go to Settings > Advanced settings and update your site configuration:

```jsonc
{
  // [...]
  "cody.enabled": true,
  "completions": {
    "enabled": true,
    "provider": "anthropic", // or "openai" if you use OpenAI
    "accessToken": "<key>",
    "endpoint": "https://api.anthropic.com/v1/complete" // or "https://api.openai.com/v1/chat/completions"
  },
  "embeddings": {
    "enabled": true,
    "provider": "openai",
    "accessToken": "<key>",
    "endpoint": "https://api.openai.com/v1/embeddings"
  },
}
```

> WARNING: When using OpenAI models for completions, only chat completions will work - code completions are currently unsupported.
