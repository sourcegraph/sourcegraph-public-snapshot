import { map } from 'rxjs/operators'

import { fetchHighlight, fetchBlobPlaintext } from '$lib/loader/blob'
import { fetchRepoCommit, queryRepositoryComparisonFileDiffs } from '$lib/loader/commits'
import { fetchLastCommit } from '$lib/repo/api/history'
import { parseRepoRevision } from '$lib/shared'
import { asStore } from '$lib/utils'

import type { PageLoad } from './$types'

export const load: PageLoad = ({ params, parent, url }) => {
    const { repoName, revision } = parseRepoRevision(params.repo)
    const compareTo = url.searchParams.get('rev')
    const compareCommit = compareTo
        ? parent().then(({ resolvedRevision }) => {
              if (!resolvedRevision) {
                  return null
              }
              return fetchRepoCommit(resolvedRevision.repo.id, compareTo)
                  .toPromise()
                  .then(result => {
                      if (result.data?.node?.__typename === 'Repository') {
                          return { commit: result.data.node.commit, repo: resolvedRevision.repo }
                      }
                      return { commit: null, repo: resolvedRevision.repo }
                  })
          })
        : null

    return {
        deferred: {
            diff: compareCommit
                ? compareCommit.then(result => {
                      if (!result?.commit?.oid || !result?.commit.parents[0]?.oid) {
                          return null
                      }
                      return queryRepositoryComparisonFileDiffs({
                          repo: result.repo.id,
                          base: result.commit?.parents[0].oid,
                          head: result.commit?.oid,
                          paths: [params.path],
                          first: null,
                          after: null,
                      }).toPromise()
                  })
                : null,
            blob: !compareTo
                ? fetchBlobPlaintext({
                      filePath: params.path,
                      repoName,
                      revision: revision ?? '',
                  }).toPromise()
                : null,
            history: compareCommit
                ? compareCommit.then(result => (result ? result.commit : null))
                : parent().then(({ resolvedRevision }) =>
                      resolvedRevision
                          ? fetchLastCommit(resolvedRevision.repo.id, resolvedRevision.commitID, params.path)
                          : null
                  ),
            highlights: compareTo
                ? Promise.resolve(null)
                : fetchHighlight({ filePath: params.path, repoName, revision: revision ?? '' })
                      .pipe(map(highlight => highlight?.lsif))
                      .toPromise(),
        },
    }
}
