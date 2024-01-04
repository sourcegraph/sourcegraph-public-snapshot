import { fetchHighlight, fetchBlobPlaintext } from '$lib/repo/api/blob'

import type { PageLoad } from './$types'
import { BlobDiffQuery } from './page.gql'

export const load: PageLoad = async ({ parent, params, url }) => {
    const revisionToCompare = url.searchParams.get('rev')
    const { resolvedRevision, graphqlClient } = await parent()

    return {
        filePath: params.path,
        deferred: {
            blob: fetchBlobPlaintext({
                filePath: params.path,
                repoID: resolvedRevision.repo.id,
                commitID: resolvedRevision.commitID,
            }),
            highlights: fetchHighlight({
                filePath: params.path,
                repoID: resolvedRevision.repo.id,
                commitID: resolvedRevision.commitID,
            }).then(highlight => highlight?.lsif),
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
        },
    }
}
