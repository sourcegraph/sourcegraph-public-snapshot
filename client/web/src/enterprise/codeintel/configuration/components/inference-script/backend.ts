import { gql } from '@sourcegraph/http-client'

export const GET_REPO_ID = gql`
    query GetRepoId($name: String!) {
        repository(name: $name) {
            id
        }
    }
`

export const INFER_JOBS_SCRIPT = gql`
    query InferAutoIndexJobsForRepo($repository: ID!, $rev: String, $script: String) {
        inferAutoIndexJobsForRepo(repository: $repository, rev: $rev, script: $script) {
            ...AutoIndexJobDescriptionFields
        }
    }

    fragment AutoIndexJobDescriptionFields on AutoIndexJobDescription {
        comparisonKey
        root
        indexer {
            key
            name
            url
        }
        steps {
            ...AutoIndexLsifIndexStepsFields
        }
    }

    fragment AutoIndexLsifIndexStepsFields on IndexSteps {
        setup {
            key
            command
        }
        preIndex {
            ...AutoIndexLsifPreIndexFields
        }
        index {
            indexerArgs
            outfile
        }
    }

    fragment AutoIndexLsifPreIndexFields on PreIndexStep {
        root
        image
        commands
    }
`
