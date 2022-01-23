import React from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useQuery, gql } from '@sourcegraph/http-client'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { SourceLocationSetWhoKnowsResult, SourceLocationSetWhoKnowsVariables } from '../../../../../graphql-operations'
import { personLinkFieldsFragment } from '../../../../../person/PersonLink'
import { ComponentOverviewWhoKnows } from '../overview/ComponentOverviewWhoKnows'

interface Props {
    sourceLocationSet: Scalars['ID']
    className?: string
}

const SOURCE_LOCATION_SET_WHO_KNOWS = gql`
    query SourceLocationSetWhoKnows($node: ID!) {
        node(id: $node) {
            __typename
            ... on SourceLocationSet {
                ...SourceLocationSetWhoKnowsFields
            }
            ... on Component {
                name
                kind
            }
        }
    }
    fragment SourceLocationSetWhoKnowsFields on SourceLocationSet {
        whoKnows {
            node {
                ...PersonLinkFields
                avatarURL
            }
            reasons
            score
        }
    }
    ${personLinkFieldsFragment}
`

export const WhoKnowsTab: React.FunctionComponent<Props> = ({ sourceLocationSet: sourceLocationSetID, className }) => {
    const { data, error, loading } = useQuery<SourceLocationSetWhoKnowsResult, SourceLocationSetWhoKnowsVariables>(
        SOURCE_LOCATION_SET_WHO_KNOWS,
        {
            variables: { node: sourceLocationSetID },
            fetchPolicy: 'cache-first',
        }
    )

    if (loading && !data) {
        return <LoadingSpinner />
    }
    if (error && !data) {
        return <ErrorAlert error={error} />
    }
    if (!data || !data.node) {
        return <ErrorAlert error="TODO(sqs)" />
    }
    if (!('whoKnows' in data.node)) {
        return <ErrorAlert error="No who-knows information" />
    }

    const sourceLocationSet = data.node

    return (
        <div className={className}>
            <ComponentOverviewWhoKnows
                whoKnows={sourceLocationSet.whoKnows}
                noun={
                    sourceLocationSet.__typename === 'Component'
                        ? `the ${sourceLocationSet.name} ${sourceLocationSet.kind.toLowerCase()}`
                        : 'the code in this directory'
                }
            />
        </div>
    )
}
