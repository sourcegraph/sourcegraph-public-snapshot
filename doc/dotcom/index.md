# Sourcegraph.com

[Sourcegraph.com](https://sourcegraph.com/search) lets you search 2 million open source repositories.

To use Sourcegraph on your own (private) code, [use Sourcegraph Cloud](../cloud/index.md) or [deploy self-hosted Sourcegraph](../admin/deploy/index.md).

- [Indexing open source code in Sourcegraph.com](indexing_open_source_code.md)

> Note: Sourcegraph dotcom indexes and allows search over a lot of code for a lot of users! For this reason sourcegraph handles some features in dotcom differently.
>
> For example, global searches do not search unindexed code by default on sourcegraph.com, whereas on a cloud or self-hosted instance this isn't the case.
> 
> To learn more about where dotcom handles things differently checkout use of the [SourcegraphDotcomMode](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+if+envvar.SourcegraphDotComMode%28%29&patternType=standard&sm=1) env var in our codebase!
