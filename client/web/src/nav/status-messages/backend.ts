import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql, dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../backend/graphql'
import { StatusMessageFields, StatusMessagesResult } from '../../graphql-operations'

export function fetchAllStatusMessages(): Observable<StatusMessageFields[]> {
    return requestGraphQL<StatusMessagesResult>(
        gql`
            query StatusMessages {
                statusMessages {
                    ...StatusMessageFields
                }
            }

            fragment StatusMessageFields on StatusMessage {
                type: __typename

                ... on CloningProgress {
                    message
                }

                ... on IndexingProgress {
                    message
                }

                ... on SyncError {
                    message
                }

                ... on IndexingError {
                    message
                }

                ... on ExternalServiceSyncError {
                    message
                    externalService {
                        id
                        displayName
                    }
                }
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.statusMessages)
    )
}
