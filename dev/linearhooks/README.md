# Linear Webhooks

## Development

> [!CAUTION]
> DO NOT commit your api key

First make a copy of the dotenv file and set the API key and webhook signing secrets in `.env` based on the [Linear API settings](https://linear.app/sourcegraph/settings/api). If you don't have access, reach out to #wg-linear-trial.

```sh
cp .env.example .env
```

```sh
source .env
```

```sh
go run .
```

Use [ngrok](https://ngrok.com/docs/getting-started/) to get a public URL for receiving webhook events:

```sh
ngrok http 3000
```

Set the [webhook URL in Linear](https://linear.app/sourcegraph/settings/api).

## Deployment

This service is deployed as a MSP service. Learn more from [go/msp](http://go/msp).

> [!CAUTION]
> Keep your secret safe

In production, it's recommended to create a [Linear OAuth2 application](https://developers.linear.app/docs/oauth/authentication), and create a developer token using application identity as actor. Then, set the developer token as `LINEAR_PERSONAL_API_KEY` in the deployment. Othewise, your personal identity will be associated with all requests.

Unfortunately, Linear only supports `authorization_code` grant type, but not `client_credentials`. Authenticating through the web interface (e.g., OAuth callback) is a lot of added complexity for a simply webhook service. We will revisit in the future.

## Configuration

Refer to [config.example.yaml](./config.example.yaml)
