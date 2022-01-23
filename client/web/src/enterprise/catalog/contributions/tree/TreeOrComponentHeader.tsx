import classNames from 'classnames'
import { LocationDescriptorObject } from 'history'
import FolderIcon from 'mdi-react/FolderIcon'
import SourceRepositoryIcon from 'mdi-react/SourceRepositoryIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { displayRepoName } from '@sourcegraph/shared/src/components/RepoFileLink'
import { Badge } from '@sourcegraph/wildcard'

import {
    PrimaryComponentForTreeFields,
    RepositoryForTreeFields,
    TreeEntryForTreeFields,
} from '../../../../graphql-operations'
import { CatalogComponentIcon } from '../../components/ComponentIcon'

import { TreeOrComponentViewOptionsProps } from './useTreeOrComponentViewOptions'

interface Props extends TreeOrComponentViewOptionsProps {
    repository: RepositoryForTreeFields
    tree: TreeEntryForTreeFields
    primaryComponent: PrimaryComponentForTreeFields | null
}

export const TreeOrComponentHeader: React.FunctionComponent<Props> = ({
    repository,
    tree,
    primaryComponent: component,
    treeOrComponentViewMode,
    treeOrComponentViewModeURL,
}) => {
    const featuredComponent = treeOrComponentViewMode === 'auto' ? component : null
    // const description = featuredComponent?.description || (tree.isRoot && repository.description) || null

    const componentFragment = component && (
        <ComponentHeading
            component={component}
            tag={featuredComponent ? 'h1' : 'h4'}
            badgeLink={featuredComponent ? undefined : treeOrComponentViewModeURL.auto}
            className="mb-2"
            textClassName={featuredComponent ? 'text-body' : 'text-muted'}
        />
    )

    return (
        <header className="mb-3 d-none">
            {!featuredComponent && componentFragment}
            <RepositoryOrTreeHeading
                repository={repository}
                tree={tree}
                tag={featuredComponent ? 'h4' : 'h1'}
                badgeLink={featuredComponent ? treeOrComponentViewModeURL.tree : undefined}
                className="mb-2"
                textClassName={featuredComponent ? 'text-muted' : 'text-body'}
            />
            {featuredComponent && componentFragment}
            {/* {description && <p className="mb-3">{description}</p>} */}
        </header>
    )
}

const RepositoryOrTreeHeading: React.FunctionComponent<{
    repository: RepositoryForTreeFields
    tree: TreeEntryForTreeFields
    tag: 'h1' | 'h4'
    badgeLink?: LocationDescriptorObject
    className?: string
    textClassName?: string
}> = ({ repository, tree, tag: Tag, badgeLink, className, textClassName }) => {
    const Icon = tree.isRoot ? SourceRepositoryIcon : FolderIcon
    const title = tree.isRoot ? displayRepoName(repository.name) : tree.path
    const to = tree.url // TODO(sqs): for repo at HEAD, should just use repo url?
    const kind = tree.isRoot ? 'repository' : 'tree'
    return (
        <Tag className={classNames('d-flex align-items-center', className)}>
            <Link to={to} className={classNames('d-flex align-items-center', textClassName)}>
                <Icon className="icon-inline mr-1" /> {title}
            </Link>
            <Badge
                variant="secondary"
                small={true}
                pill={true}
                className={classNames('ml-1', textClassName)}
                as={LinkOrSpan}
                to={badgeLink}
            >
                {kind}
            </Badge>
        </Tag>
    )
}

const ComponentHeading: React.FunctionComponent<{
    component: PrimaryComponentForTreeFields
    tag: 'h1' | 'h4'
    badgeLink?: LocationDescriptorObject
    className?: string
    textClassName?: string
}> = ({ component, tag: Tag, badgeLink, className, textClassName }) => (
    <Tag className={classNames('d-flex align-items-center', className)}>
        <Link to={component.url} className={classNames('d-flex align-items-center', textClassName)}>
            <CatalogComponentIcon component={component} className="icon-inline mr-2 flex-shrink-0" /> {component.name}
        </Link>
        <Badge
            variant="secondary"
            small={true}
            pill={true}
            className={classNames('ml-1', textClassName)}
            as={LinkOrSpan}
            to={badgeLink}
        >
            {component.kind.toLowerCase()}
        </Badge>
    </Tag>
)
