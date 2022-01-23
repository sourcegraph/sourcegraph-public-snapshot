import EmailIcon from 'mdi-react/EmailIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { useQuery, gql } from '@sourcegraph/http-client'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import {
    SourceSetWhoKnowsFields,
    SourceSetWhoKnowsResult,
    SourceSetWhoKnowsVariables,
} from '../../../../../graphql-operations'
import { personLinkFieldsFragment } from '../../../../../person/PersonLink'
import { PersonList } from '../../../components/person-list/PersonList'

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
            <WhoKnows
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

const WhoKnows: React.FunctionComponent<{
    whoKnows: SourceSetWhoKnowsFields['whoKnows']
    noun: string
}> = ({ whoKnows, noun }) => (
    <PersonList
        title={`Who knows about ${noun}?`}
        description={
            <p className="text-muted small mb-2">
                Suggestions are automatically generated based on code contributions, ownership, and usage.
            </p>
        }
        listTag="ol"
        orientation="vertical"
        primaryText="person"
        items={whoKnows.map(({ node: person, score, reasons }) => ({
            person,
            text: (
                <ul className="list-inline">
                    {reasons.map((reason, index) => (
                        <li key={reason} className="list-inline-item">
                            {index !== 0 && <span className="mr-2">&bull;</span>}
                            {reason}
                        </li>
                    ))}
                </ul>
            ),
            textTooltip: score.toFixed(1),
            action: (
                <>
                    <a
                        // TODO(sqs): hacky
                        href={`https://slack.com/app_redirect?channel=@${person.email.slice(
                            0,
                            person.email.indexOf('@')
                        )}`}
                        target="_blank"
                        rel="noopener"
                        className="btn btn-secondary btn-sm mr-2"
                    >
                        <SlackIcon className="icon-inline" /> @{person.email.slice(0, person.email.indexOf('@'))}
                    </a>
                    <a href={`mailto:${person.email}`} className="btn btn-secondary btn-sm">
                        <EmailIcon className="icon-inline" /> Email
                    </a>
                </>
            ),
        }))}
        listClassName="card border-0"
    />
)
