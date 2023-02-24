// We want to limit the number of imported modules as much as possible
/* eslint-disable no-restricted-imports */

import type { Observable } from 'rxjs'
import { map, tap } from 'rxjs/operators'

import { getWebGraphQLClient, mutateGraphQL, requestGraphQL } from '@sourcegraph/web/src/backend/graphql'

import { resetAllMemoizationCaches } from './common'
import type { CheckMirrorRepositoryConnectionResult, Scalars } from './graphql-operations'
import { dataOrThrowErrors, gql } from './http-client'

export type { ResolvedRevision } from '@sourcegraph/web/src/repo/backend'
export { syntaxHighlight } from '@sourcegraph/web/src/repo/blob/codemirror/highlight'
export { parseSearchURL } from '@sourcegraph/web/src/search/index'
export { defaultSearchModeFromSettings } from '@sourcegraph/web/src/util/settings'
export { replaceRevisionInURL } from '@sourcegraph/web/src/util/url'
export { requestGraphQL, getWebGraphQLClient, mutateGraphQL }

// Copy of non-reusable code

// Importing from @sourcegraph/web/site-admin/backend.ts breaks the build because this
// module has (transitive) dependencies on @sourcegraph/wildcard which imports
// all Wildcard components
//
const CHECK_MIRROR_REPOSITORY_CONNECTION = gql`
    mutation CheckMirrorRepositoryConnection($repository: ID, $name: String) {
        checkMirrorRepositoryConnection(repository: $repository, name: $name) {
            error
        }
    }
`
export function checkMirrorRepositoryConnection(
    args:
        | {
              repository: Scalars['ID']
          }
        | {
              name: string
          }
): Observable<CheckMirrorRepositoryConnectionResult['checkMirrorRepositoryConnection']> {
    return mutateGraphQL<CheckMirrorRepositoryConnectionResult>(CHECK_MIRROR_REPOSITORY_CONNECTION, args).pipe(
        map(dataOrThrowErrors),
        tap(() => resetAllMemoizationCaches()),
        map(data => data.checkMirrorRepositoryConnection)
    )
}
