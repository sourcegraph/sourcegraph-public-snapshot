import React from 'react'
import { Link } from 'react-router-dom'
import { HighlightedMatches } from '../../../../../../../shared/src/components/HighlightedMatches'
import { QueryParameterProps } from '../../../../../components/withQueryParameter/WithQueryParameter'

interface Props extends Pick<QueryParameterProps, 'query'> {
    icon: React.ComponentType<{ className?: string }>
    title: string
    count: number
    className?: string
}

/**
 * An item in the thread changes sidebar's filter list.
 */
export const ThreadChangesSidebarFilterListItem: React.FunctionComponent<Props> = ({
    icon: Icon,
    title,
    count,
    query,
    className = '',
}) => (
    <Link to="#" className={`d-flex align-items-center ${className}`} title={title}>
        <Icon className="icon-inline small mr-1 flex-0" />
        <span className="flex-1 text-truncate mr-1">
            <HighlightedMatches text={title} pattern={query} />
        </span>
        <span className="flex-0">{count}</span>
    </Link>
)
