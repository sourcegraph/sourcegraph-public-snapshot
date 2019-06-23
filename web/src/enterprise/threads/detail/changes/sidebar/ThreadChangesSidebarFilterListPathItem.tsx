import FileIcon from 'mdi-react/FileIcon'
import React from 'react'
import { QueryParameterProps } from '../../../../../components/withQueryParameter/WithQueryParameter'
import { ThreadChangesSidebarFilterListItem } from './ThreadChangesSidebarFilterListItem'

interface Props extends Pick<QueryParameterProps, 'query'> {
    path: string
    count: number

    className?: string
}

/**
 * A path item in the thread changes sidebar's filter list.
 */
export const ThreadChangesSidebarFilterListPathItem: React.FunctionComponent<Props> = ({ path, count, ...props }) => (
    <ThreadChangesSidebarFilterListItem {...props} icon={FileIcon} title={path} count={count} />
)
