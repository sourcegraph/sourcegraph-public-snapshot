import classNames from 'classnames'
import React from 'react'

import { CatalogEntityWhoKnowsFields } from '../../../../../graphql-operations'

import { PersonList } from './PersonList'

interface Props {
    catalogComponent: CatalogEntityWhoKnowsFields
    className?: string
}

export const EntityWhoKnowsTab: React.FunctionComponent<Props> = ({ catalogComponent: { whoKnows }, className }) => (
    <div className={classNames(className)}>
        <div className="container my-3">
            <PersonList
                title="Who knows"
                listTag="ol"
                orientation="vertical"
                primaryText="person"
                items={whoKnows.map(({ node: person, score, reasons }) => ({
                    person,
                    text: (
                        <ul className="list-inline">
                            {reasons.map((reason, index) => (
                                <li key={index} className="list-inline-item">
                                    {index !== 0 && <span className="mr-2">&bull;</span>}
                                    {reason}
                                </li>
                            ))}
                        </ul>
                    ),
                    textTooltip: score.toFixed(1),
                }))}
                className={className}
            />
        </div>
    </div>
)
