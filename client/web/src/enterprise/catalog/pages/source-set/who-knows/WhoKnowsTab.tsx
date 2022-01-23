import React from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useQuery, gql } from '@sourcegraph/http-client'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import { SourceSetWhoKnowsResult, SourceSetWhoKnowsVariables } from '../../../../../graphql-operations'
import { personLinkFieldsFragment } from '../../../../../person/PersonLink'
import { ComponentOverviewWhoKnows } from '../overview/ComponentOverviewWhoKnows'

interface Props {
    sourceSet: Scalars['ID']
    className?: string
}

const SOURCE_SET_WHO_KNOWS = gql`
    query SourceSetWhoKnows($node: ID!) {
        node(id: $node) {
            __typename
            ... on SourceSet {
                ...SourceSetWhoKnowsFields
            }
            ... on Component {
                name
                kind
            }
        }
    }
    fragment SourceSetWhoKnowsFields on SourceSet {
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

export const WhoKnowsTab: React.FunctionComponent<Props> = ({ sourceSet: sourceSetID, className }) => {
    const { data, error, loading } = useQuery<SourceSetWhoKnowsResult, SourceSetWhoKnowsVariables>(
        SOURCE_SET_WHO_KNOWS,
        {
            variables: { node: sourceSetID },
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

    const sourceSet = data.node

    return (
        <div className={className}>
            <ComponentOverviewWhoKnows
                whoKnows={sourceSet.whoKnows}
                noun={
                    sourceSet.__typename === 'Component'
                        ? `the ${sourceSet.name} ${sourceSet.kind.toLowerCase()}`
                        : 'the code in this directory'
                }
            />
        </div>
    )
}
