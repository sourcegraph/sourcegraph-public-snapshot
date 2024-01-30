import type { PageLoad } from './$types'
import { BlobDiffQuery, BlobPageQuery, BlobSyntaxHighlightQuery } from './page.gql'

export const load: PageLoad = async ({ parent, params, url }) => {
    const revisionToCompare = url.searchParams.get('rev')
    const { resolvedRevision, graphqlClient } = await parent()

    return {
        filePath: params.path,
        blob: graphqlClient
            .query({
                query: BlobPageQuery,
                variables: {
                    repoID: resolvedRevision.repo.id,
                    revspec: resolvedRevision.commitID,
                    path: params.path,
                },
            })
            .then(result => {
                if (result.data.node?.__typename !== 'Repository' || !result.data.node.commit?.blob) {
                    throw new Error('Commit or file not found')
                }
                return result.data.node.commit.blob
            }),
        highlights: graphqlClient
            .query({
                query: BlobSyntaxHighlightQuery,
                variables: {
                    repoID: resolvedRevision.repo.id,
                    revspec: resolvedRevision.commitID,
                    path: params.path,
                    disableTimeout: false,
                },
            })
            .then(result => {
                if (result.data.node?.__typename !== 'Repository') {
                    throw new Error('Expected Repository')
                }
                return result.data.node.commit?.blob?.highlight.lsif
            }),
        compare: revisionToCompare
            ? {
                  revisionToCompare,
                  diff: graphqlClient
                      .query({
                          query: BlobDiffQuery,
                          variables: {
                              repoID: resolvedRevision.repo.id,
                              revspec: revisionToCompare,
                              paths: [params.path],
                          },
                      })
                      .then(result => {
                          if (result.data.node?.__typename !== 'Repository') {
                              throw new Error('Expected Repository')
                          }
                          return result.data.node.commit?.diff.fileDiffs.nodes[0]
                      }),
              }
            : null,
    }
}
