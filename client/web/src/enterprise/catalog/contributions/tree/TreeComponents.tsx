import classNames from 'classnames'
import FolderIcon from 'mdi-react/FolderIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { useQuery, gql } from '@sourcegraph/http-client'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { FileSpec } from '@sourcegraph/shared/src/util/url'

import {
    OtherComponentForTreeEntryFields,
    PrimaryComponentForTreeEntryFields,
    ComponentsForTreeEntryResult,
    ComponentsForTreeEntryVariables,
} from '../../../../graphql-operations'
import { pathHasPrefix, pathRelative } from '../../../../util/path'
import { CatalogComponentIcon } from '../../components/ComponentIcon'
import { COMPONENT_AUTHORS_FRAGMENT } from '../../pages/component/gql'
import { COMPONENT_OWNER_FRAGMENT } from '../../pages/component/meta/ComponentOwnerSidebarItem'

import styles from './TreeComponents.module.scss'

interface Props extends FileSpec {
    repoID: Scalars['ID']
    className?: string
}

const COMPONENTS_FOR_TREE_ENTRY = gql`
    query ComponentsForTreeEntry($repository: ID!, $path: String!) {
        node(id: $repository) {
            __typename
            ... on Repository {
                ...ComponentsForTreeEntryFields
            }
        }
    }

    fragment ComponentsForTreeEntryFields on Repository {
        primaryComponents: components(path: $path, primary: true, recursive: false) {
            ...PrimaryComponentForTreeEntryFields
        }
        otherComponents: components(path: $path, recursive: true) {
            ...OtherComponentForTreeEntryFields
        }
    }

    fragment PrimaryComponentForTreeEntryFields on Component {
        __typename
        id
        name
        kind
        description
        lifecycle
        url
        ...ComponentOwnerFields
    }

    fragment OtherComponentForTreeEntryFields on Component {
        __typename
        id
        name
        kind
        description
        url
        sourceLocations {
            repository {
                id
            }
            path
            isEntireRepository
            treeEntry {
                url
            }
            isPrimary
        }
    }

    ${COMPONENT_OWNER_FRAGMENT}
`

export const TreeComponents: React.FunctionComponent<Props> = ({ repoID, filePath, className }) => {
    // TODO(sqs): use commitID to adjust forward so that you see components when browsing old, since-renamed paths?

    const { data, error, loading } = useQuery<ComponentsForTreeEntryResult, ComponentsForTreeEntryVariables>(
        COMPONENTS_FOR_TREE_ENTRY,
        {
            variables: { repository: repoID, path: filePath },
            fetchPolicy: 'cache-first',
        }
    )

    if (error) {
        throw error
    }
    if (loading && !data) {
        return null
    }

    const primaryComponents = (data && data.node?.__typename === 'Repository' && data.node.primaryComponents) || null
    const otherComponents =
        (data &&
            data.node?.__typename === 'Repository' &&
            data.node.otherComponents.filter(component => component.id !== primaryComponents?.[0]?.id)) ||
        null

    if ((!otherComponents || otherComponents.length === 0) && (!primaryComponents || primaryComponents.length === 0)) {
        return null
    }

    return (
        <section className={className}>
            {primaryComponents && primaryComponents.length > 0 && (
                <ComponentDetail
                    component={primaryComponents[0]}
                    className={classNames('px-3 pt-3 border border-secondary rounded', styles.componentDetail)}
                />
            )}
            {otherComponents && otherComponents.length > 0 && (
                <ul className={classNames('list-unstyled', styles.boxGrid)}>
                    {otherComponents.map(component => (
                        <ComponentGridItem
                            key={component.id}
                            component={component}
                            treeRepoID={repoID}
                            treePath={filePath}
                            className={classNames('border border-secondary rounded', styles.boxGridItem)}
                            linkBigClickAreaClassName={styles.linkBigClickArea}
                        />
                    ))}
                </ul>
            )}
        </section>
    )
}

const ComponentDetail: React.FunctionComponent<{
    component: PrimaryComponentForTreeEntryFields
    className?: string
}> = ({ component, className }) => (
    <div className={className}>
        <h3>
            <Link to={component.url} className="d-flex align-items-center font-weight-bold">
                <CatalogComponentIcon component={component} className="icon-inline mr-1 text-muted" /> {component.name}
            </Link>
        </h3>
        <div className="text-muted small mb-2">
            {component.__typename === 'Component' && `${component.kind[0]}${component.kind.slice(1).toLowerCase()}`}
        </div>
        {component.description && <p>{component.description}</p>}
    </div>
)

const ComponentGridItem: React.FunctionComponent<{
    component: OtherComponentForTreeEntryFields
    treeRepoID: Scalars['ID']
    treePath: string
    tag?: 'li'
    className?: string
    linkBigClickAreaClassName?: string
}> = ({ component, treeRepoID, treePath, tag: Tag = 'li', className, linkBigClickAreaClassName }) => {
    const primarySourceLocation = component.sourceLocations.find(({ isPrimary }) => isPrimary)
    if (!primarySourceLocation) {
        throw new Error('unable to determine primary source location')
    }

    const nearestSourceLocation = component.sourceLocations.find(
        sourceLocation =>
            sourceLocation.repository?.id === treeRepoID &&
            (sourceLocation.path === null || pathHasPrefix(sourceLocation.path, treePath))
    )
    if (!nearestSourceLocation) {
        throw new Error('unable to determine nearest source location')
    }

    return (
        <Tag className={classNames('d-flex flex-column', className)}>
            <div className="position-relative flex-1">
                <h4 className="mb-0">
                    <Link
                        to={component.url}
                        className={classNames(
                            'd-flex align-items-center stretched-link font-weight-bold',
                            linkBigClickAreaClassName
                        )}
                    >
                        <CatalogComponentIcon component={component} className="icon-inline mr-1 text-muted" />
                        {component.name}
                    </Link>
                </h4>
                {component.description && (
                    <p className={classNames('my-1 small', styles.boxGridItemBody)}>{component.description}</p>
                )}
            </div>
            <ul className="list-unstyled small mt-1">
                <SourceLocationItem
                    sourceLocation={nearestSourceLocation}
                    treePath={treePath}
                    linkBigClickAreaClassName={
                        primarySourceLocation === nearestSourceLocation ? linkBigClickAreaClassName : undefined
                    }
                />
                {primarySourceLocation !== nearestSourceLocation && (
                    <SourceLocationItem
                        sourceLocation={primarySourceLocation}
                        treePath={treePath}
                        linkBigClickAreaClassName={linkBigClickAreaClassName}
                    />
                )}
            </ul>
        </Tag>
    )
}

const SourceLocationItem: React.FunctionComponent<{
    sourceLocation: OtherComponentForTreeEntryFields['sourceLocations'][0]
    treePath: string
    className?: string
    linkBigClickAreaClassName?: string
}> = ({ sourceLocation, treePath, className, linkBigClickAreaClassName }) => (
    <li className={className}>
        <LinkOrSpan
            to={sourceLocation.treeEntry?.url /* TODO(sqs): this takes you away from the current rev back to HEAD */}
            className={classNames('d-flex align-items-center text-muted', linkBigClickAreaClassName)}
        >
            <FolderIcon className="icon-inline mr-1 flex-shrink-0" />
            <span className="text-truncate">
                {treePath === sourceLocation.path
                    ? 'This directory'
                    : pathRelative(treePath, sourceLocation.path || '/')}
            </span>
        </LinkOrSpan>
    </li>
)
