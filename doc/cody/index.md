# Cody (experimental)

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This feature is experimental and might change or be removed in the future. We've released it as an experimental feature to provide a preview of functionality we're working on.
</p>
</aside>

Cody is an AI coding assistant that lives in your editor that can find, explain, and write code. Cody uses a combination of Large Language Models (LLMs), Sourcegraph search, and Sourcegraph code intelligence to provide answers that eliminate toil and keep human programmers in flow. You can think of Cody as your programmer buddy who has read through all the code in open source, all the questions on StackOverflow, and all your organization's private code, and is always there to answer questions you might have or suggest ways of doing something based on prior knowledge.

Cody is in private alpha (tagged as an [experimental](../admin/beta_and_experimental_features.md) feature) at this stage. 
- If you are an existing Sourcegraph Enterprise customer or want to use Cody for your team, contact your techical advisor or [sign up here](https://sourcegraph.typeform.com/to/pIXTgwrd) to get access
- If you want to try Cody on open source code, sign up [here](https://forms.gle/cffMa8mrr8YuHv8o8) and we'll e-mail you instructions to connect Cody to sourcegraph.com as soon as your account is added.

Currently, Cody is available for VS Code. More editors are on the way—[join the Discord](https://discord.gg/8wJF5EdAyA) to inquire about your editor of choice.

## Cody on Sourcegraph.com

Cody uses Sourcegraph to fetch relevant context to generate answers and code. These instructions walk through installing Cody and connecting it to sourcegraph.com. For private instances of Sourcegraph, see the section below about enabling Cody for Enterprise.

1. Sign into [sourcegraph.com](https://sourcegraph.com)
1. Request access [here](https://forms.gle/cffMa8mrr8YuHv8o8) and we'll send you an e-mail you as soon as your account is added. At this time, we are approving all requests.
1. [Create a Sourcegraph access token](https://sourcegraph.com/user/settings/tokens)
1. Install [the Cody VS Code extension](https://marketplace.visualstudio.com/items?itemName=sourcegraph.cody-ai)
  1. Set the Sourcegraph URL to be `https://sourcegraph.com`
  1. Set the access token to be the token you just created
1. See [this section](#enabling-codebase-aware-answers) on how to enable codebase-aware answers.

<img width="553" alt="Cody login screen" src="https://user-images.githubusercontent.com/25070988/227510233-5ce37649-6ae3-4470-91d0-71ed6c68b7ef.png">

After installing, we recommend the following:

* [See the list](embedded-repos.md) of embedded repositories and request any that you'd like to add by pinging a Sourcegraph team member in [Discord](https://discord.gg/8wJF5EdAyA). Embeddings significantly improve the accuracy and quality of Cody's responses. Note that embeddings are only available for public repositories on sourcegraph.com. If you want to use Cody with embeddings on private code, consider moving to a Sourcegraph Enterprise instance.
* Spread the word online and send us your feedback in Discord. Cody is open source and we'd love to hear from you if you have bug reports or feature requests.

## Cody on Sourcegraph Cloud

On Sourcegraph Cloud, Cody is a managed service and you do not need to follow the step 1 of the self-hosted guide. 

1. Cody can be enabled on demand on your Sourcegraph instance by contacting your account manager. The Sourcegraph team will refer to the [handbook](https://handbook.sourcegraph.com/departments/cloud/#managed-instance-requests).
1. Users can then configure the [VS Code extension](#step-2-configure-the-vs-code-extension)

Learn more from [Cody on Cloud](../cloud/index.md#cody).

## Cody on your self-hosted Sourcegraph Enterprise instance

### Prerequisites

- Sourcegraph 5.0.1 or above.
- An Anthropic API key, that you can get from your Technical Advisor or Customer Engineer. 
- (Optionally), an OpenAI API key for embeddings.

There are two steps required to enable Cody for Enterprise: enable your Sourcegraph instance and configure the VS Code extension.

### Step 1: Enable Cody on your Sourcegraph instance

Note that this requires site-admin privileges.

1. Cody uses one or more third-party LLM (Large Language Model) providers. Make sure you review the [Cody usage and privacy notice](https://about.sourcegraph.com/terms/cody-notice). In particular, code snippets will be sent to a third-party language model provider when you use the Cody extension.
2. To turn Cody on, you will need to set an access token for Sourcegraph to authentify with the third-party large language model provider (currently Anthropic but we may use different or several models over time). Reach out to your Sourcegraph Technical Advisor or Customer Engineer to get a key. They will create a key for you using the [anthropic console](https://console.anthropic.com/account/keys).
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
5. Cody can be configured to use embeddings to significantly improve the quality of its responses. This involves sending your entire codebase to a third-party service to generate a low-dimensional semantic representation, that is used for improved context fetching. See the [embeddings](#embeddings) section for more.

### Step 2: Configure the VS Code extension

Now that Cody is turned on on your Sourcegraph instance, any user can configure and use the Cody VS Code extension. This does not require admin privilege.

1. If you currently have a previous version of Cody installed, uninstall it and reload VS Code before proceeding to the next steps.
1. Search for “Sourcegraph Cody” in your VS Code extension marketplace, and install it.

<img width="500" alt="Sourcegraph Cody in VS Code Marketplace" src="https://user-images.githubusercontent.com/55068936/228114612-65402e1c-7501-44cb-a846-46c4376b9572.png">

3. Reload VS Code, and open the Cody extension. Review and accept the terms.

4. Now you'll need to point the Cody extension to your Sourcegraph instance. On your instance, go to `settings` / `access token` (`https://<your-instance>.sourcegraph.com/users/<your-instance>/settings/tokens`). Generate an access token, copy it, and set it in the Cody extension.

<img width="1369" alt="image" src="https://user-images.githubusercontent.com/25070988/227510686-4afcb1f9-a3a5-495f-b1bf-6d661ba53cce.png">

5. In the Cody VS Code extension, set your instance URL and the access token
    
<img width="553" alt="image" src="https://user-images.githubusercontent.com/25070988/227510233-5ce37649-6ae3-4470-91d0-71ed6c68b7ef.png">

6. See [this section](#enabling-codebase-aware-answers) on how to enable codebase-aware answers.

You're all set!

### Step 3: Try Cody!

A few things you can ask Cody:

- "What are popular go libraries for building CLIs?"
- Open your workspace, and ask "Do we have a React date picker component in this repository?"
- Right click on a function, and ask Cody to explain it
- Try any of the Cody recipes!

<img width="510" alt="image" src="https://user-images.githubusercontent.com/25070988/227511383-aa60f074-817d-4875-af41-54558dfe1951.png">

## Enabling codebase-aware answers

The `Cody: Codebase` setting in VS Code enables codebase-aware answers for the Cody extension. By setting this configuration option to the repository name on your Sourcegraph instance, Cody will be able to provide more accurate and relevant answers to your coding questions, based on the context of the codebase you are currently working in.

1. Open the VS Code workspace settings by pressing <kbd>Cmd/Ctrl+,</kbd>, (or File > Preferences (Settings) on Windows & Linux).
2. Search for the `Cody: Codebase` setting.
3. Enter the repository name as listed on your Sourcegraph instance.
   1. For example: `github.com/sourcegraph/sourcegraph` without the `https` protocol

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
  "dimensions": 1536,
  "excludedFilePathPatterns": []
}
```

* Navigate to Site admin > Cody (`/site-admin/cody`) and schedule repositories for embedding.

> NOTE: By enabling Cody, you agree to the [Cody Notice and Usage Policy](https://about.sourcegraph.com/terms/cody-notice). 

### Excluding files from embeddings

The `excludedFilePathPatterns` is a setting in the Sourcegraph embeddings configuration that allows you to exclude certain file paths from being used in generating embeddings. By specifying glob patterns that match file paths, you can exclude files that have low information value, such as test fixtures, mocks, auto-generated files, and other files that are not relevant to the codebase.

To use `excludedFilePathPatterns`, add it to your embeddings site config with a list of glob patterns. For example, to exclude all SVG files, you would add the following setting to your configuration file:

```json
"embeddings": {
  // ...
  "excludedFilePathPatterns": [
    "*.svg"
  ]
}
```

> NOTE: The `excludedFilePathPatterns` setting is only available in Sourcegraph version `5.0.1` and later.

### Storing embedding indexes

To target a managed object storage service, you will need to set a handful of environment variables for configuration and authentication to the target service. **If you are running a sourcegraph/server deployment, set the environment variables on the server container. Otherwise, if running via Docker-compose or Kubernetes, set the environment variables on the `frontend`, `embeddings`, and `worker` containers.**

#### Using S3

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


#### Using GCS

To target a GCS bucket you've already provisioned, set the following environment variables. Authentication is done through a service account key, supplied as either a path to a volume-mounted file, or the contents read in as an environment variable payload.

- `EMBEDDINGS_UPLOAD_BACKEND=GCS`
- `EMBEDDINGS_UPLOAD_BUCKET=<my bucket name>`
- `EMBEDDINGS_UPLOAD_GCP_PROJECT_ID=<my project id>`
- `EMBEDDINGS_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE=</path/to/file>`
- `EMBEDDINGS_UPLOAD_GOOGLE_APPLICATION_CREDENTIALS_FILE_CONTENT=<{"my": "content"}>`

#### Provisioning buckets

If you would like to allow your Sourcegraph instance to control the creation and lifecycle configuration management of the target buckets, set the following environment variables:

- `EMBEDDINGS_UPLOAD_MANAGE_BUCKET=true`

### Environment variables for the `embeddings` service

- `EMBEDDINGS_REPO_INDEX_CACHE_SIZE`: Number of repository embedding indexes to cache in memory (the default cache size is 5). Increasing the cache size will improve the search performance but require more memory resources.


## Turning Cody off

To turn Cody off, set `embeddings` and `completions` site-admin settings to `enabled:false` (or remove them altogether).

