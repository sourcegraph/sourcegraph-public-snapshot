import { gql } from '@sourcegraph/http-client'

const ingestedCodeownersFragment = gql`
    fragment IngestedCodeowners on CodeownersIngestedFile {
        contents
        updatedAt
    }
`

export const GET_INGESTED_CODEOWNERS_QUERY = gql`
    ${ingestedCodeownersFragment}
    query GetIngestedCodeowners($repoID: ID!) {
        node(id: $repoID) {
            ... on Repository {
                ingestedCodeowners {
                    ...IngestedCodeowners
                }
            }
        }
    }
`

export const ADD_INGESTED_CODEOWNERS_MUTATION = gql`
    ${ingestedCodeownersFragment}
    mutation AddIngestedCodeowners($repoID: ID!, $contents: String!) {
        addCodeownersFile(input: { repoID: $repoID, fileContents: $contents }) {
            ...IngestedCodeowners
        }
    }
`

export const UPDATE_INGESTED_CODEOWNERS_MUTATION = gql`
    ${ingestedCodeownersFragment}
    mutation UpdateIngestedCodeowners($repoID: ID!, $contents: String!) {
        updateCodeownersFile(input: { repoID: $repoID, fileContents: $contents }) {
            ...IngestedCodeowners
        }
    }
`

export const DELETE_INGESTED_CODEOWNERS_MUTATION = gql`
    mutation DeleteIngestedCodeowners($repoID: ID!) {
        deleteCodeownersFiles(repositories: [{ repoID: $repoID }]) {
            alwaysNil
        }
    }
`
