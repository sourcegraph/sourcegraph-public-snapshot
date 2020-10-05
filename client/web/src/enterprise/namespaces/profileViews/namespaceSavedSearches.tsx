import React from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { View, ViewContexts } from '../../../../../shared/src/api/client/services/viewService'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../../backend/graphql'
import { NamespaceSavedSearchesResult, NamespaceSavedSearchesVariables } from '../../../graphql-operations'
import { pluralize } from '../../../../../shared/src/util/strings'
import { ContributableViewContainer } from '../../../../../shared/src/api/protocol'

export const namespaceSavedSearches = ({
    id:_id,
}: ViewContexts[typeof ContributableViewContainer.Profile]): Observable<View | null> => {
    const savedSearches = requestGraphQL<NamespaceSavedSearchesResult, NamespaceSavedSearchesVariables>(
        gql`
            query NamespaceSavedSearches {
                savedSearches {
                    id
                    description
                    query
                    notify
                }
            }
        `,
        {  }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.savedSearches)
    )

    return savedSearches.pipe(
        map(savedSearches => ({
            title: `${savedSearches.length} ${pluralize('saved search', savedSearches.length, 'saved searches')}`,
            titleLink: 'TODO',
            content: [
                {
                    reactComponent: () => <div>Saved searches: ${JSON.stringify(savedSearches)}</div>,
                },
            ],
        }))
    )
}
