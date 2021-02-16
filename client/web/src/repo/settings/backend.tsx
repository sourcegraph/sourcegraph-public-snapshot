import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { RepoNotFoundError } from '../../../../shared/src/backend/errors'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../backend/graphql'
import {
    SettingsAreaRepositoryFields,
    SettingsAreaRepositoryResult,
    SettingsAreaRepositoryVariables,
} from '../../graphql-operations'

export const settingsAreaRepositoryFragment = gql`
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
                id
                kind
                displayName
            }
        }
        permissionsInfo {
            syncedAt
            updatedAt
        }
    }
`

/**
 * Fetches a repository.
 */
export function fetchSettingsAreaRepository(name: string): Observable<SettingsAreaRepositoryFields> {
    return requestGraphQL<SettingsAreaRepositoryResult, SettingsAreaRepositoryVariables>(
        gql`
            query SettingsAreaRepository($name: String!) {
                repository(name: $name) {
                    ...SettingsAreaRepositoryFields
                }
            }
            ${settingsAreaRepositoryFragment}
        `,
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
