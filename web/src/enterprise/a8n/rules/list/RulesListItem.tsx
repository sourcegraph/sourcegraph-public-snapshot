import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../../shared/src/graphql/schema'

interface Props {
    rule: GQL.IRule

    tag: 'li'
    className?: string
}

/**
 * A single rule in a list.
 */
export const RulesListItem: React.FunctionComponent<Props> = ({ rule, tag: Tag, className = '' }) => (
    <Tag className={`d-flex flex-wrap align-items-stretch position-relative ${className}`}>
        <h3 className="mb-0 font-weight-normal font-size-base d-flex align-items-center">
            <Link to={rule.url} className="stretched-link">
                {rule.name}
            </Link>
        </h3>
    </Tag>
)
