import FileIcon from 'mdi-react/FileIcon'
import React from 'react'
import { QueryParameterProps } from '../../../components/withQueryParameter/WithQueryParameter'
import { ThreadInboxSidebarFilterListItem } from './ThreadInboxSidebarFilterListItem'

interface Props extends QueryParameterProps {
    path: string
    count: number

    className?: string
}

/**
 * A path item in the thread inbox sidebar's filter list.
 */
export const ThreadInboxSidebarFilterListPathItem: React.FunctionComponent<Props> = ({ path, count, ...props }) => (
    <ThreadInboxSidebarFilterListItem {...props} icon={FileIcon} title={path} count={count} />
)
