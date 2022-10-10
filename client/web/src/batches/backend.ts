import { gql } from '@sourcegraph/http-client'

/**
 * NOTE: These fields are only available from an enterprise install, but are used to
 * surface batch changes information on the repo `TreePage`, which is available to both
 * OSS and enterprise
 */
export const REPO_CHANGESETS_STATS = gql`
    query RepoChangesetsStats($name: String!) {
        repository(name: $name) {
            changesetsStats {
                open
                merged
            }
        }
    }
`
