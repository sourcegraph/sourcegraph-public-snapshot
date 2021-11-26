import EmailIcon from 'mdi-react/EmailIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React from 'react'

import { ComponentStateDetailFields, SourceLocationSetWhoKnowsFields } from '../../../../../graphql-operations'
import { PersonList } from '../PersonList'

interface Props {
    whoKnows: SourceLocationSetWhoKnowsFields['whoKnows']
    noun: string
}

export const ComponentOverviewWhoKnows: React.FunctionComponent<Props> = ({ whoKnows, noun }) => (
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

export function whoKnowsDescription(entity: Pick<ComponentStateDetailFields, 'name' | 'kind'>): string {
    return `Need help with the ${entity.name} ${entity.kind.toLowerCase()}? See who knows about it.`
}
