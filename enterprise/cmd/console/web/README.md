# web

## Running locally

```
caddy run
yarn run esbuild:serve
```

## Configuration

### Google OAuth app

First, create a new Google Cloud project for the tasklist instance's OAuth credentials. (This project can be different from where it's deployed.)

```shell
PROJECT=sourcegraph-tasklist
gcloud projects create $PROJECT
gcloud config set project $PROJECT
gcloud services enable docs.googleapis.com drive.googleapis.com tasks.googleapis.com
```

Next, create the Google OAuth 2.0 app and credentials:

1. [Create an OAuth consent screen](https://console.cloud.google.com/apis/credentials/consent) for the app. (Ensure that the correct Google Cloud project is selected.)
   - **User Type:** Internal
   - **App name:** (your choice)
   - **User support email:** (your choice)
   - **Authorized domains:** (the domain where the app is deployed, such as `example.com`)
   - **Developer contact information:** (your choice)
   - **Scopes:** Click **Add or remove scopes**, and paste the following under **Manually add scopes**: `https://www.googleapis.com/auth/userinfo.email,https://www.googleapis.com/auth/drive.readonly,https://www.googleapis.com/auth/documents.readonly,https://www.googleapis.com/auth/tasks.readonly`
1. [Create an OAuth 2.0 Client ID](https://console.cloud.google.com/apis/credentials) for the app.
   - Click **Create Credentials > OAuth client ID**
   - **Application type:** Web application
   - **Name:** (your choice)
   - **Authorized JavaScript origins:** (the origin where the app is deployed, such as `https://tasklist.example.com:1234`)

Use the newly generated client ID TODO(sqs).

Also create an API key:

1. [Create an API key](https://console.cloud.google.com/apis/credentials) for the app.
   - Click **Create Credentials > API key**
   - TODO sqs set restrictions for referer etc
