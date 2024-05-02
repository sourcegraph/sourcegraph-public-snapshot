# Linear Webhooks

## Development

Make of a copy of the dotenv file. You can retrieve Linear API key and webhook signing secrets from [Linear settings](https://linear.app/sourcegraph/settings/api). If you don't have access, reach out to #wg-linear-trial.

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

## Configuration

Refer to [config.example.yaml](./config.example.yaml)
