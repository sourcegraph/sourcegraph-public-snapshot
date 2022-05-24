const cachedContentRequests = new Map<string, Promise<string>>()

import { ContentMatch } from '@sourcegraph/shared/src/search/stream'

import { getIdForMatch } from '../results/utils'

export async function loadContent(match: ContentMatch): Promise<string> {
    const cacheKey = getIdForMatch(match)

    if (cachedContentRequests.has(cacheKey)) {
        return (await cachedContentRequests.get(cacheKey)) as string
    }

    const loadPromise = fetchBlobContent(match)
    cachedContentRequests.set(cacheKey, loadPromise)

    // When the content fails to load, remove the cache entry to allow reloading
    loadPromise.catch(() => cachedContentRequests.delete(cacheKey))

    return loadPromise
}

async function fetchBlobContent(match: ContentMatch): Promise<string> {
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-explicit-any
    const response: any = await fetch('https://sourcegraph.com/.api/graphql', {
        method: 'post',
        body: JSON.stringify({
            query: `
                query Blob($repoName: String!, $commitID: String!, $filePath: String!) {
                    repository(name: $repoName) {
                        commit(rev: $commitID) {
                            file(path: $filePath) {
                                content
                            }
                        }
                    }
                }`,
            variables: {
                commitID: match.commit,
                filePath: match.path,
                repoName: match.repository,
            },
        }),
    }).then(response => response.json())
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment,@typescript-eslint/no-unsafe-member-access
    const content: undefined | string = response?.data?.repository?.commit?.file?.content
    if (content === undefined) {
        throw new Error('No content found in query response')
    }
    return content
}
