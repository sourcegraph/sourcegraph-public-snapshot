# Cody FAQ

### General

#### Troubleshooting

See [Cody troubleshooting guide](troubleshooting.md).

#### Does Cody train on my code?

Cody doesn't train on your code. Our third-party LLM providers do not train on your code either.

The way Cody generates an answer is the following:

- A user asks a question.
- Sourcegraph uses the code intelligence platform (search, code intelligence, embeddings) to retrieve code relevant to the question. In that process, permissions are enforced and only code that the user has read permission on is retrieved.
- Sourcegraph sends a prompt to an LLM to answer, providing the code retrieved as context.
- The reply is sent to Cody.

#### Does Cody work with self-hosted Sourcegraph?

Yes, Cody works with self-hosted Sourcegraph instances, with the caveat that snippets of code (up to 28 KB per request) will be sent to a third party cloud service (Anthropic by default, but can also be OpenAI) on each request. Optionally, embeddings can be turned on for some repositories, which requires sending those repositories to another third party (OpenAI).

In particular, this means the Sourcegraph instance needs to be able to access the internet.

#### Does Cody require Sourcegraph to function?

Yes. Sourcegraph is needed both to retrieve context and as a proxy for the LLM provider.

### Embeddings

#### What are embeddings for?

Embeddings are one of the many ways Sourcegraph uses to retrieve relevant code to feed the large language model as context. Embeddings / vector search are complementary to other strategies. While it matches really well semantically ("what is this code about, what does it do?"), it drops syntax and other important precise matching info. Sourcegraph's overall approach is to blend results from multiple sources to provide the best answer possible.


#### When using embeddings, are permissions enforced? Does Cody get fed code that the users doesn't have access to?

Permissions are enforced when using embeddings. Today, Sourcegraph only uses embeddings search on a single repo, first checking that the users has access.

In the future, here are the steps that Sourcegraph will follow:

- determine which repo you have access to
- query embeddings for each of those repo
- pick the best results and send it back

### Third party dependencies

#### What third-party cloud services does Cody depend on today?

- Cody has one third-party dependency, which is Anthropic's Claude API. In the config, this can be replaced with OpenAI API.
- Cody can optionally use OpenAI to generate embeddings, that are then used to improve the quality of its context snippets, but this is not required.

#### What's the retention policy for Anthropic/OpenAI?

See our [terms](https://about.sourcegraph.com/terms/cody-notice).

#### Can I use my own API keys?

Yes!

