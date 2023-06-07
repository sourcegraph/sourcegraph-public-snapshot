import { isErrorLike } from '@sourcegraph/common'

import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'

import { UnifiedContextFetcher, UnifiedContextFetcherResult } from '.'

export class UnifiedContextFetcherClient implements UnifiedContextFetcher {
    constructor(private client: SourcegraphGraphQLAPIClient, private repoIds: string[]) {}

    public async getContext(
        query: string,
        codeResultsCount: number,
        textResultsCount: number
    ): Promise<UnifiedContextFetcherResult[] | Error> {
        const response = await this.client.getCodyContext(this.repoIds, query, codeResultsCount, textResultsCount)

        if (isErrorLike(response)) {
            return response
        }

        return response.reduce((results, result) => {
            if (result) {
                results.push({
                    filePath: result.blob.path,
                    content: result.chunkContent,
                    startLine: result.startLine,
                    endLine: result.endLine,
                    repoName: result.blob.repository.name,
                    revision: result.blob.commit.oid,
                })
            }

            return results
        }, [] as UnifiedContextFetcherResult[])
    }
}
