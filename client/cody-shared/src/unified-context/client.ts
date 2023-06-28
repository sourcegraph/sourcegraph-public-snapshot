import { isErrorLike } from '@sourcegraph/common'

import { ContextSearchOptions } from '../codebase-context'
import { SourcegraphGraphQLAPIClient } from '../sourcegraph-api/graphql'

import { UnifiedContextFetcher, UnifiedContextFetcherResult } from '.'

export class UnifiedContextFetcherClient implements UnifiedContextFetcher {
    constructor(private client: SourcegraphGraphQLAPIClient, private repoIds: string[]) {}

    public async getContext(
        query: string,
        options: ContextSearchOptions
    ): Promise<UnifiedContextFetcherResult[] | Error> {
        const response = await this.client.getCodyContext(this.repoIds, query, options)

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
                    commit:
                        options.includeOwnershipContext && result.blob.commit.author
                            ? {
                                  id: result.blob.commit.id,
                                  oid: result.blob.commit.oid,
                                  date: result.blob.commit.author.date,
                                  author: result.blob.commit.author.person.name,
                                  subject: result.blob.commit.subject || '',
                              }
                            : undefined,
                    owner:
                        options.includeOwnershipContext && result.blob.ownership?.nodes.length
                            ? {
                                  reason: result.blob.ownership.nodes[0].reasons[0]?.__typename as any,
                                  type: result.blob.ownership.nodes[0].owner.__typename as any,
                                  name: result.blob.ownership.nodes[0].owner.name,
                              }
                            : undefined,
                })
            }

            return results
        }, [] as UnifiedContextFetcherResult[])
    }
}
