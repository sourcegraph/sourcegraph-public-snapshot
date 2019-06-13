import FileIcon from 'mdi-react/FileIcon'
import React from 'react'
import { RepositoryIcon } from '../../../../../../../shared/src/components/icons'
import { displayRepoName } from '../../../../../../../shared/src/components/RepoFileLink'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { QueryParameterProps } from '../../../components/withQueryParameter/WithQueryParameter'
import { ThreadInboxSidebarFilterListItem } from './ThreadInboxSidebarFilterListItem'

interface Props extends QueryParameterProps {
    repository: Pick<GQL.IRepository, 'name'>
    count: number

    className?: string
}

/**
 * A repository item in the thread inbox sidebar's filter list.
 */
export const ThreadInboxSidebarFilterListRepositoryItem: React.FunctionComponent<Props> = ({
    repository,
    count,
    ...props
}) => (
    <ThreadInboxSidebarFilterListItem
        {...props}
        icon={RepositoryIcon}
        title={displayRepoName(repository.name)}
        count={count}
    />
)
