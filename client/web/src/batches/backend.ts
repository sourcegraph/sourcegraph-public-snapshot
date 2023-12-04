import { gql } from '@sourcegraph/http-client'

export const REPO_CHANGESETS_STATS = gql`
    query RepoChangesetsStats($name: String!) {
        repository(name: $name) {
            id
            changesetsStats {
                open
                merged
            }
        }
    }
`
