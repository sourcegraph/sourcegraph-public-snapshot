import { gql } from '@sourcegraph/http-client'

export const REPO_EMBEDDING_EXISTS_QUERY = gql`
    query RepoEmbeddingExistsQuery($repoName: String!) {
        repository(name: $repoName) {
            id
            embeddingExists
        }
    }
`
