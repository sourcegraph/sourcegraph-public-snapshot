# Why am I getting a result count mismatch in Batch vs Web

This document seeks to explain a result count mismatch when running a search query in batch and when running the same query in the web UI.

## Prerequisites
This document assumes that you have a running instance of Sourcegraph either on Kubernetes or Docker.

## Explanation
When running a query in both web and in the batch CLI, different results don’t mean that batch or web is not correctly updating all locations, what it does mean is:
In the web UI, results refer to ‘matches’. For context, you are able to get x matches because each result contains multiple text matches.
The web UI also gets back some duplicate results sometimes as a filepath and as a text match.

In conclusion, this shouldn’t be considered as a block for you as you run search queries.

## Further resources
[Sourcegraph Code Search](https://docs.sourcegraph.com/code_search)
