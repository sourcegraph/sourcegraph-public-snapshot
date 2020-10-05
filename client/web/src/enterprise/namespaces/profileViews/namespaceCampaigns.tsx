import React from 'react'
import { Observable } from 'rxjs'
import { filter, map } from 'rxjs/operators'
import { View, ViewContexts } from '../../../../../shared/src/api/client/services/viewService'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../../backend/graphql'
import { NamespaceCampaignsResult, NamespaceCampaignsVariables } from '../../../graphql-operations'
import { pluralize } from '../../../../../shared/src/util/strings'
import { ContributableViewContainer } from '../../../../../shared/src/api/protocol'
import { isDefined } from '../../../../../shared/src/util/types'

export const namespaceCampaigns = ({
    id,
}: ViewContexts[typeof ContributableViewContainer.Profile]): Observable<View | null> => {
    const campaigns = requestGraphQL<NamespaceCampaignsResult, NamespaceCampaignsVariables>(
        gql`
            query NamespaceCampaigns($id: ID!, $first: Int!) {
                node(id: $id) {
                    ... on User {
                        campaigns(first: $first) {
                            ...NamespaceCampaignsFields
                        }
                    }
                    ... on Org {
                        campaigns(first: $first) {
                            ...NamespaceCampaignsFields
                        }
                    }
                }
            }

            fragment NamespaceCampaignsFields on CampaignConnection {
                nodes {
                    namespace {
                        namespaceName
                    }
                    name
                    url
                }
                totalCount
            }
        `,
        { id, first: 5 }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node?.campaigns)
    )

    return campaigns.pipe(
        filter(isDefined),
        map(campaigns => ({
            title: `${campaigns.totalCount} ${pluralize('campaign', campaigns.totalCount)}`,
            titleLink: 'TODO',
            content: [
                {
                    reactComponent: () => <div>Campaigns: ${JSON.stringify(campaigns)}</div>,
                },
            ],
        }))
    )
}
