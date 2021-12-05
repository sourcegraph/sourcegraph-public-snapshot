import classNames from 'classnames'
import EmailIcon from 'mdi-react/EmailIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { CatalogEntityDetailFields, CatalogEntityWhoKnowsFields } from '../../../../../graphql-operations'

import { PersonList } from './PersonList'

interface Props {
    catalogComponent: CatalogEntityWhoKnowsFields & Pick<CatalogEntityDetailFields, 'name' | 'kind'>
    className?: string
}

export const EntityWhoKnowsTab: React.FunctionComponent<Props> = ({
    catalogComponent: { whoKnows, ...entity },
    className,
}) => (
    <div className={classNames(className)}>
        <div className="container my-3">
            <PersonList
                title={`Who knows about the ${entity.name} ${entity.kind.toLowerCase()}?`}
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
                            <Link to={`mailto:${person.email}`} className="btn btn-secondary btn-sm mr-2">
                                <EmailIcon className="icon-inline" /> Email
                            </Link>
                            <Link to={`mailto:${person.email}`} className="btn btn-secondary btn-sm ">
                                <SlackIcon className="icon-inline" /> Slack
                            </Link>
                        </>
                    ),
                }))}
                listClassName="card border-0"
            />
        </div>
    </div>
)

export function whoKnowsDescription(entity: Pick<CatalogEntityDetailFields, 'name' | 'kind'>): string {
    return `Need help with the ${entity.name} ${entity.kind.toLowerCase()}? See who knows about it.`
}
