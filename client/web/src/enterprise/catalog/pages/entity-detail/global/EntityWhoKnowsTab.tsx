import classNames from 'classnames'
import React from 'react'

import { CatalogComponentWhoKnowsFields } from '../../../../../graphql-operations'

import { PersonList } from './PersonList'

interface Props {
    catalogComponent: CatalogComponentWhoKnowsFields
    className?: string
}

export const EntityWhoKnowsTab: React.FunctionComponent<Props> = ({ catalogComponent: { whoKnows }, className }) => (
    <div className={classNames(className)}>
        <div className="container my-3">
            <PersonList
                title="Who knows"
                listTag="ol"
                orientation="vertical"
                items={whoKnows.slice(0, 10).map(({ node: person, score, reasons }) => ({
                    person,
                    text: reasons.join(', '),
                    textTooltip: score.toFixed(1),
                }))}
                className={className}
            />
        </div>
    </div>
)
