# Sourcegraph Cody Gateway

<span class="badge badge-note">Sourcegraph 5.1+</span>

Sourcegraph Cody Gateway powers the default `"provider": "sourcegraph"` Cody completions and embeddings for Sourcegraph Enterprise customers.
It supports a variety of upstream LLM providers, such as [Anthropic](https://www.anthropic.com/) and [OpenAI](https://openai.com/), with rate limits, quotas, and model availability tied to your Sourcegraph Enterprise product subscription.
[Sourcegraph App users](./../overview/app/index.md) with Sourcegraph.com accounts will also be able to use Sourcegraph Cody Gateway.

Reach out your account manager for more details about Sourcegraph Cody Gateway access available to you and how you can gain access to higher rate limits, quotas, and/or model options.

## Using Cody Gateway in Sourcegraph Enterprise

> WARNING: Sourcegraph Cody Gateway access must be included in your Sourcegraph Enterprise subscription plan first - reach out to your account manager for more details.
>
> If you are a [Sourcegraph Cloud](../../cloud/index.md) customer, Cody is enabled by default on your instance starting with Sourcegraph 5.1.

<span class="virtual-br"></span>

> NOTE: Sourcegraph Cody Gateway uses one or more third-party LLM (Large Language Model) providers. Make sure you review the [Cody usage and privacy notice](https://about.sourcegraph.com/terms/cody-notice). In particular, code snippets will be sent to a third-party language model provider when you use the Cody extension or when embeddings are enabled.

To enable completions and embeddings provided by Cody Gateway on your Sourcegraph Enterprise instance, simply ensure your license key is set and Cody is enabled in [site configuration](../../admin/config/site_config.md):

```jsonc
{
  "licenseKey": "<...>",
  "cody.enabled": true,
}
```

That's it! Reasonable defaults will automatically be applied, and authentication will happen automatically based on the configured license key.

For more details about configuring Cody, refer to the following guides:

- [Enabling Cody for Sourcegraph Enterprise](./../overview/enable-cody-enterprise.md)
- [Code Graph Context: Embeddings](./code_graph_context.md#embeddings)

Cody Gateway is hosted at `cody-gateway.sourcegraph.com`. To use Cody Gateway, your Sourcegraph instance must be able to connect to the service at this domain.

## Configuring custom models

To configure custom models for various Cody configurations (e.g. `"completions"` and `"embeddings"`), specify the desired model with the upstream provider as a prefix to the name of the model. For example, to use the `claude-2` model from Anthropic, you would configure:

```json
{
  "completions": { "chatModel": "anthropic/claude-2" },
}
```

The currently supported upstream providers for models are:

- [`anthropic/`](https://www.anthropic.com/)
- [`openai/`](https://openai.com/)

For Sourcegraph Enterprise customers, model availability depends on your Sourcegraph Enterprise subscription - reach out your account manager for more details.

Refer to [Cody documentation](../overview/index.md) to learn more about Cody configuration.

> WARNING: When using OpenAI models for completions, only chat completions will work - code completions are currently unsupported.

## Rate limits and quotas

Rate limits, quotas, and model availability is tied to one of:

- your Sourcegraph Enterprise product subscription, for Sourcegraph Enterprise instances
- your Sourcegraph.com account, for [Sourcegraph App users](../overview/app/index.md)

All successful requests to Cody Gateway will count towards your rate limits.
Unsuccesful requests are not counted as usage.
Rate limits, quotas, and model availability are also configured per Cody feature - for example, you will have a separate rate limits for Cody chat, Cody completions, and Cody embeddings.

In addition to the above, we may throttle concurrent requests to Cody Gateway per Sourcegraph Enterprise instance or Sourcegraph App user, to prevent excessive burst consumption.

## Privacy and security

Sourcegraph Cody Gateway does not retain any sensitive data (prompt test and source code included in requests, etc) from any traffic we receive.
We only tracks rate limit consumption per Sourcegraph Enterprise subscription, and some high-level diagnostic data (errors codes from upstream, numeric/enum request parameters, etc).
The code that powers Cody Gateway is also [source-available](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph$+f:cmd/cody-gateway+lang:go&patternType=lucky&sm=1&groupBy=path) for audit.

For more details about Cody Gateway security practices, please reach out to your account manager.
Also refer to the [Cody usage and privacy notice](https://about.sourcegraph.com/terms/cody-notice) for more privacy details about Cody in general.
