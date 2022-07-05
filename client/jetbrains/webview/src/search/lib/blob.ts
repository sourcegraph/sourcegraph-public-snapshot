import { gql } from '@sourcegraph/http-client'
import { ContentMatch, PathMatch, SymbolMatch } from '@sourcegraph/shared/src/search/stream'

import { BlobContentResult, BlobContentVariables } from '../../graphql-operations'
import { getMatchId } from '../results/utils'

import { requestGraphQL } from './requestGraphQl'

const cachedContentRequests = new Map<string, Promise<string | null>>()

export async function loadContent(match: ContentMatch | PathMatch | SymbolMatch): Promise<string | null> {
    const cacheKey = getMatchId(match)

    if (cachedContentRequests.has(cacheKey)) {
        return (await cachedContentRequests.get(cacheKey)) as string
    }

    const loadPromise = fetchBlobContent(match)
    cachedContentRequests.set(cacheKey, loadPromise)

    // When the content fails to load, remove the cache entry to allow reloading
    loadPromise.catch(() => cachedContentRequests.delete(cacheKey))

    return loadPromise
}

async function fetchBlobContent(match: ContentMatch | PathMatch | SymbolMatch): Promise<string | null> {
    const response = await requestGraphQL<BlobContentResult, BlobContentVariables>(blobContentQuery, {
        commitID: match.commit ?? '',
        filePath: match.path,
        repoName: match.repository,
    })

    const content: undefined | string = response.data?.repository?.commit?.file?.content
    if (content === undefined) {
        console.error('No content found in query response', response)
        return null
    }
    return content
}

const blobContentQuery = gql`
    query BlobContent($repoName: String!, $commitID: String!, $filePath: String!) {
        repository(name: $repoName) {
            commit(rev: $commitID) {
                file(path: $filePath) {
                    content
                    # We include the highlight part here even though it is not used to get a server side
                    # error when previewing binary files.
                    highlight(disableTimeout: false) {
                        aborted
                    }
                }
            }
        }
    }
`
