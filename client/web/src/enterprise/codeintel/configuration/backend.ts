import { gql } from '@sourcegraph/http-client'

export const GET_REPO_ID = gql`
    query GetRepoId($name: String!) {
        repository(name: $name) {
            id
        }
    }
`

export const AutoIndexJobFieldsFragment = gql`
    fragment AutoIndexJobDescriptionFields on AutoIndexJobDescription {
        comparisonKey
        root
        indexer {
            key
            imageName
            name
            url
        }
        steps {
            ...AutoIndexLsifIndexStepsFields
        }
    }

    fragment AutoIndexLsifIndexStepsFields on IndexSteps {
        preIndex {
            ...AutoIndexLsifPreIndexFields
        }
        index {
            ...AutoIndexLsifIndexFields
        }
    }

    fragment AutoIndexLsifPreIndexFields on PreIndexStep {
        root
        image
        commands
    }

    fragment AutoIndexLsifIndexFields on IndexStep {
        indexerArgs
        outfile
        commands
        requestedEnvVars
    }
`

export const INFER_JOBS_SCRIPT = gql`
    query InferAutoIndexJobsForRepo($repository: ID!, $rev: String, $script: String) {
        inferAutoIndexJobsForRepo(repository: $repository, rev: $rev, script: $script) {
            jobs {
                ...AutoIndexJobDescriptionFields
            }
            inferenceOutput
        }
    }

    ${AutoIndexJobFieldsFragment}
`

export const INFERRED_CONFIGURATION = gql`
    query InferredIndexConfiguration($id: ID!) {
        node(id: $id) {
            ...RepositoryInferredIndexConfigurationFields
        }
    }

    fragment RepositoryInferredIndexConfigurationFields on Repository {
        __typename
        indexConfiguration {
            inferredConfiguration {
                configuration
                parsedConfiguration {
                    ...AutoIndexJobDescriptionFields
                }
            }
        }
    }

    ${AutoIndexJobFieldsFragment}
`

export const REPOSITORY_CONFIGURATION = gql`
    query IndexConfiguration($id: ID!) {
        node(id: $id) {
            ...RepositoryIndexConfigurationFields
        }
    }

    fragment RepositoryIndexConfigurationFields on Repository {
        __typename
        indexConfiguration {
            configuration
            parsedConfiguration {
                ...AutoIndexJobDescriptionFields
            }
        }
    }

    ${AutoIndexJobFieldsFragment}
`
