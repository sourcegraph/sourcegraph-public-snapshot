// We want to limit the number of imported modules as much as possible
/* eslint-disable no-restricted-imports */

import { gql, mutation } from './graphql'
import type { CheckMirrorRepositoryConnectionResult, Scalars } from './graphql-operations'

export { parseSearchURL } from '@sourcegraph/web/src/search/index'
export { replaceRevisionInURL } from '@sourcegraph/web/src/util/url'

export { syntaxHighlight } from '@sourcegraph/web/src/repo/blob/codemirror/highlight'
export {
    selectableLineNumbers,
    type SelectedLineRange,
    setSelectedLines,
} from '@sourcegraph/web/src/repo/blob/codemirror/linenumbers'
export { isValidLineRange } from '@sourcegraph/web/src/repo/blob/codemirror/utils'
export { blobPropsFacet } from '@sourcegraph/web/src/repo/blob/codemirror'
export { defaultSearchModeFromSettings } from '@sourcegraph/web/src/util/settings'
export { GlobalNotebooksArea, type GlobalNotebooksAreaProps } from '@sourcegraph/web/src/notebooks/GlobalNotebooksArea'
export {
    CodeInsightsRouter,
    type CodeInsightsRouterProps,
} from '@sourcegraph/web/src/enterprise/insights/CodeInsightsRouter'

export type { FeatureFlagName } from '@sourcegraph/web/src/featureFlags/featureFlags'

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
): Promise<CheckMirrorRepositoryConnectionResult['checkMirrorRepositoryConnection']> {
    return mutation<CheckMirrorRepositoryConnectionResult>(CHECK_MIRROR_REPOSITORY_CONNECTION, args).then(
        data => data?.checkMirrorRepositoryConnection ?? { error: null }
    )
}
