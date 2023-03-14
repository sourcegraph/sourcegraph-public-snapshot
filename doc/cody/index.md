# Cody (experimental)

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This feature is experimental and might change or be removed in the future. We've released it as an experimental feature to provide a preview of functionality we're working on.
</p>
</aside>

Cody is an AI coding assistant that lives in your editor that can find, explain, and write code. Cody uses a combination of AI (specifically Large Language Models or LLMs), Sourcegraph search, and Sourcegraph code intelligence to provide answers that eliminate toil and keep human programmers in flow. You can think of Cody as your programmer buddy who has read through all the code on GitHub, all the questions on StackOverflow, and all your organization's private code, and is always there to answer questions you might have or suggest ways of doing something based on prior knowledge.

Cody is in private alpha (tagged as an [experimental](../admin/beta_and_experimental_features.md) feature) at this stage. Contact your techical advisor or [signup here](https://t.co/4TMTW1b3lR) to get access. We have limited capacity to onboard customers at first, but we're working hard to open Cody up to the world fast! In this initial release, Cody is only available as a VS Code extension.

To enable the Cody extension to work with your Sourcegraph instance (requires site-admin permissions), follow the steps below. Note that this assumes you've reached out to us to get required `accessToken`(s).

* Go to Site admin > Site configuration (`/site-admin/configuration`) on your instance.
* (Required) Add the `completions` config:

```json
"completions": {
  "enabled": true,
  "accessToken": "<token>",
  "model": "claude-v1",
  "provider": "anthropic"
}
```

* (Optional) Add the `embeddings` config:

```json
"embeddings": {
  "enabled": true,
  "url": "",
  "accessToken": "<token>",
  "model": "",
  "dimensions": 128
}
```

Here is the config for the OpenAI Embeddings API:

```json
"embeddings": {
  "enabled": true,
  "url": "https://api.openai.com/v1/embeddings",
  "accessToken": "<token>",
  "model": "text-embedding-ada-002",
  "dimensions": 1536
}
```

* Navigate to Site admin > Cody (`/site-admin/cody`) and schedule repositories for embedding.

> NOTE: By enabling Cody, you agree to the [Cody Notice and Usage Policy](https://about.sourcegraph.com/terms/cody-notice). In particular, some code snippets will be sent to a third-party language model provider when you use the Cody extension.

## Embeddings

Embeddings are a semantic representation of text. Embeddings are usually floating-point vectors with 256+ elements. The useful thing about embeddings is that they allow us to search over textual information using a semantic correlation between the query and the text, not just syntactic (matching keywords). We are using embeddings to create a search index over an entire codebase which allows us to perform natural language code search over the codebase. Indexing involves splitting the **entire codebase** into searchable chunks, and sending them to the external service specified in the site config for embedding. The final embedding index is stored in a managed object storage service. The available storage configurations are listed in the next section.

## Storing embedding indexes

To target a managed object storage service, you will need to set a handful of environment variables for configuration and authentication to the target service. **If you are running a sourcegraph/server deployment, set the environment variables on the server container. Otherwise, if running via Docker-compose or Kubernetes, set the environment variables on the `frontend`, `embeddings`, and `worker` containers.**

### Using S3

To target an S3 bucket you've already provisioned, set the following environment variables. Authentication can be done through [an access and secret key pair](https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys) (and optional session token), or via the EC2 metadata API.

**_Warning:_** Remember never to commit aws access keys in git. Consider using a secret handling service offered by your cloud provider. 

- `EMBEDDINGS_UPLOAD_BACKEND=S3`
- `EMBEDDINGS_UPLOAD_BUCKET=<my bucket name>`
- `EMBEDDINGS_UPLOAD_AWS_ENDPOINT=https://s3.us-east-1.amazonaws.com`
- `EMBEDDINGS_UPLOAD_AWS_ACCESS_KEY_ID=<your access key>`
- `EMBEDDINGS_UPLOAD_AWS_SECRET_ACCESS_KEY=<your secret key>`
- `EMBEDDINGS_UPLOAD_AWS_SESSION_TOKEN=<your session token>` (optional)
- `EMBEDDINGS_UPLOAD_AWS_USE_EC2_ROLE_CREDENTIALS=true` (optional; set to use EC2 metadata API over static credentials)
- `EMBEDDINGS_UPLOAD_AWS_REGION=us-east-1` (default)

**_Note:_** If a non-default region is supplied, ensure that the subdomain of the endpoint URL (_the `AWS_ENDPOINT` value_) matches the target region.

> NOTE: You don't need to set the `EMBEDDINGS_UPLOAD_AWS_ACCESS_KEY_ID` environment variable when using `EMBEDDINGS_UPLOAD_AWS_USE_EC2_ROLE_CREDENTIALS=true` because role credentials will be automatically resolved. 


### Using GCS

To target a GCS bucket you've already provisioned, set the following environment variables. Authentication is done through a service account key, supplied as either a path to a volume-mounted file, or the contents read in as an environment variable payload.

- `EMBEDDINGS_UPLOAD_BACKEND=GCS`
- `EMBEDDINGS_UPLOAD_BUCKET=<my bucket name>`
- `EMBEDDINGS_UPLOAD_GCP_PROJECT_ID=<my project id>`
- `EMBEDDINGS_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE=</path/to/file>`
- `EMBEDDINGS_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE_CONTENT=<{"my": "content"}>`

### Provisioning buckets

If you would like to allow your Sourcegraph instance to control the creation and lifecycle configuration management of the target buckets, set the following environment variables:

- `EMBEDDINGS_UPLOAD_MANAGE_BUCKET=true`
