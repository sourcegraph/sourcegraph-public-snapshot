import { memoize } from 'lodash'

import { CodebaseContext } from '@sourcegraph/cody-shared/src/codebase-context'
import { SourcegraphEmbeddingsSearchClient } from '@sourcegraph/cody-shared/src/embeddings/client'
import { KeywordContextFetcher } from '@sourcegraph/cody-shared/src/keyword-context'
import { isError } from '@sourcegraph/cody-shared/src/utils'

import { sourcegraphClient } from './sourcegraph-client'

/**
 * Memoized function to get the repository ID for a given codebase.
 */
const getRepoId = memoize(async (codebase: string) => {
    const repoId = codebase ? await sourcegraphClient.getRepoId(codebase) : null

    if (isError(repoId)) {
        const errorMessage =
            `Cody could not find the '${codebase}' repository on your Sourcegraph instance.\n` +
            'Please check that the repository exists and is entered correctly in the cody.codebase setting.'
        console.error(errorMessage)
    }

    return repoId
})

export async function createCodebaseContext(
    codebase: string,
    contextType: 'embeddings' | 'keyword' | 'none' | 'blended'
) {
    const repoId = await getRepoId(codebase)
    const embeddingsSearch =
        repoId && !isError(repoId) ? new SourcegraphEmbeddingsSearchClient(sourcegraphClient, repoId) : null

    const codebaseContext = new CodebaseContext(
        { useContext: contextType },
        embeddingsSearch,
        new LocalKeywordContextFetcherMock()
    )

    return codebaseContext
}

class LocalKeywordContextFetcherMock implements KeywordContextFetcher {
    public getContext() {
        return Promise.resolve([])
    }
}
