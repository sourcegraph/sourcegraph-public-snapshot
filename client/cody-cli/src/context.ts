import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { SourcegraphEmbeddingsSearchClient } from '@sourcegraph/cody-shared/src/embeddings/client'
import { KeywordContextFetcher } from '@sourcegraph/cody-shared/src/keyword-context'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { isError } from '@sourcegraph/cody-shared/src/utils'

const getRepoId = async (client: SourcegraphGraphQLAPIClient, codebase: string) => {
    const repoId = codebase ? await client.getRepoId(codebase) : null
    return repoId
}

export async function createCodebaseContext(
    client: SourcegraphGraphQLAPIClient,
    codebase: string,
    contextType: 'embeddings' | 'keyword' | 'none' | 'blended',
    serverEndpoint: string
) {
    const repoId = await getRepoId(client, codebase)
    if (isError(repoId)) {
        throw repoId
    }

    const embeddingsSearch = repoId && !isError(repoId) ? new SourcegraphEmbeddingsSearchClient(client, repoId) : null

    const codebaseContext = new CodebaseContext(
        { useContext: contextType, serverEndpoint },
        codebase,
        embeddingsSearch,
        new LocalKeywordContextFetcherMock()
    )

    return codebaseContext
}

class LocalKeywordContextFetcherMock implements KeywordContextFetcher {
    public getContext() {
        return Promise.resolve([])
    }
    public getSearchContext() {
        return Promise.resolve([])
    }
}
