# Cody (experimental)

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This feature is experimental and might change or be removed in the future. We've released it as an experimental feature to provide a preview of functionality we're working on.
</p>
</aside>

Cody is an AI coding assistant that lives in your editor that can find, explain, and write code. Cody uses a combination of AI (specifically Large Language Models or LLMs), Sourcegraph search, and Sourcegraph code intelligence to provide answers that eliminate toil and keep human programmers in flow. You can think of Cody as your programmer buddy who has read through all the code on GitHub, all the questions on StackOverflow, and all your organization's private code, and is always there to answer questions you might have or suggest ways of doing something based on prior knowledge.

Cody is in private alpha (tagged as an [experimental](../doc/admin/beta_and_experimental_features.md) feature) at this stage. Contact your techical advisor or [signup here](https://t.co/4TMTW1b3lR) to get access.

So far, Cody is available only for VS Code.

## How to enable Cody

There are two steps required to enable Cody: enabling Cody on your Sourcegraph instance, and configuring the VS Code extension.

### Step 1: Enable Cody on your Sourcegraph instance

Note that this requires site-admin privileges.

1. Cody uses one or more third-party LLM (Large Language Model) providers. Make sure you review the [Cody usage and privacy notice](https://about.sourcegraph.com/terms/cody-notice). In particular, code snippets will be sent to a third-party language model provider when you use the Cody extension.
2. To turn Cody on, you will need to set an access token for Sourcegraph to authentify with the third-party large language model provider (currently Anthropic but we may use different or several models over time). Reach out to your Sourcegraph Technical Advisor to get a key.
3. Once you have the key, go to Site admin > Site configuration (`/site-admin/configuration`) on your instance and set:

```json
"completions": {
  "enabled": true,
  "accessToken": "<token>",
  "model": "claude-v1",
  "provider": "anthropic"
}
```
4. You're done! 
5. (Optional). Cody can be configured to use embeddings to improve the quality of its responses. This involves sending your entire codebase to a third-party service to generate a low-dimensional semantic representation, that is used for improved context fetching. See the [embeddings](#embeddings) section for more.

### Step 2: Configure the VS Code extension

Now that Cody is turned on on your Sourcegraph instance, any user can configure and use the Cody VS Code extension. This does not require admin privilege.

1. If you currently have a previous version of Cody installed, uninstall it and reload VS Code before proceeding to the next steps.
1. Search for “Cody Enterprise” in your VS Code extension marketplace, and install it.

<img width="500" alt="image" src="https://user-images.githubusercontent.com/25070988/227508342-cc6f29c0-ed91-4381-b651-16870c7c676b.png">

3. Reload VS Code, and open the Cody extension. Review and accept the terms.

<img width="687" alt="image" src="https://user-images.githubusercontent.com/25070988/227509535-600f7d96-a241-4f5d-a89f-399e6c0f9019.png">

4. Now you'll need to point the Cody extension to your Sourcegraph instance. On your instance, go to `settings` / `access token` (https://<your-instance>.sourcegraph.com/users/<your-instance>/settings/tokens). Generate an access token, copy it, and set it in the Cody extension.

<img width="1369" alt="image" src="https://user-images.githubusercontent.com/25070988/227510686-4afcb1f9-a3a5-495f-b1bf-6d661ba53cce.png">

5. In the Cody VS Code extension, set your instance URL and the access token
    
<img width="553" alt="image" src="https://user-images.githubusercontent.com/25070988/227510233-5ce37649-6ae3-4470-91d0-71ed6c68b7ef.png">

You're all set!


## Embeddings

Embeddings are a semantic representation of text. Embeddings are usually floating-point vectors with 256+ elements. The useful thing about embeddings is that they allow us to search over textual information using a semantic correlation between the query and the text, not just syntactic (matching keywords). We are using embeddings to create a search index over an entire codebase which allows us to perform natural language code search over the codebase. Indexing involves splitting the **entire codebase** into searchable chunks, and sending them to the external service specified in the site config for embedding. The final embedding index is stored in a managed object storage service. The available storage configurations are listed in the next section.

### Configuring embeddings

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

> NOTE: By enabling Cody, you agree to the [Cody Notice and Usage Policy](https://about.sourcegraph.com/terms/cody-notice). 

### Storing embedding indexes

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
