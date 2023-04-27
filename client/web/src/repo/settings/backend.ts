import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { RepoNotFoundError } from '@sourcegraph/shared/src/backend/errors'

import { requestGraphQL } from '../../backend/graphql'
import {
    SettingsAreaRepositoryFields,
    SettingsAreaRepositoryResult,
    SettingsAreaRepositoryVariables,
} from '../../graphql-operations'

export const settingsAreaRepositoryFragment = gql`
    fragment SettingsAreaExternalServiceFields on ExternalService {
        id
        kind
        displayName
        supportsRepoExclusion
    }

    fragment SettingsAreaRepositoryFields on Repository {
        id
        name
        url
        isPrivate
        viewerCanAdminister
        mirrorInfo {
            remoteURL
            cloneInProgress
            cloneProgress
            cloned
            updatedAt
            isCorrupted
            corruptionLogs {
                timestamp
                reason
            }
            lastError
            updateSchedule {
                due
                index
                total
            }
            updateQueue {
                updating
                index
                total
            }
        }
        externalServices {
            nodes {
                ...SettingsAreaExternalServiceFields
            }
        }
        metadata {
            key
            value
        }
    }
`

export const FETCH_SETTINGS_AREA_REPOSITORY_GQL = gql`
    query SettingsAreaRepository($name: String!) {
        repository(name: $name) {
            ...SettingsAreaRepositoryFields
        }
    }
    ${settingsAreaRepositoryFragment}
`

export const ADD_REPO_METADATA_GQL = gql`
    mutation AddRepoMetadata($repo: ID!, $key: String!, $value: String) {
        addRepoMetadata(repo: $repo, key: $key, value: $value) {
            alwaysNil
        }
    }
`

export const DELETE_REPO_METADATA_GQL = gql`
    mutation DeleteRepoMetadata($repo: ID!, $key: String!) {
        deleteRepoMetadata(repo: $repo, key: $key) {
            alwaysNil
        }
    }
`

/**
 * Fetches a repository.
 */
export function fetchSettingsAreaRepository(name: string): Observable<SettingsAreaRepositoryFields> {
    return requestGraphQL<SettingsAreaRepositoryResult, SettingsAreaRepositoryVariables>(
        FETCH_SETTINGS_AREA_REPOSITORY_GQL,
        { name }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.repository) {
                throw new RepoNotFoundError(name)
            }
            return data.repository
        })
    )
}

export const EXCLUDE_REPO_FROM_EXTERNAL_SERVICES = gql`
    mutation ExcludeRepoFromExternalServices($externalServices: [ID!]!, $repo: ID!) {
        excludeRepoFromExternalServices(externalServices: $externalServices, repo: $repo) {
            alwaysNil
        }
    }
`
