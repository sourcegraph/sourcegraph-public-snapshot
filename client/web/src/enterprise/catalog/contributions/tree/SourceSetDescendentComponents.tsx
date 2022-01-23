import classNames from 'classnames'
import FolderIcon from 'mdi-react/FolderIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { gql } from '@sourcegraph/http-client'
import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { FileSpec } from '@sourcegraph/shared/src/util/url'

import { SourceSetDescendentComponentsFields } from '../../../../graphql-operations'
import { pathHasPrefix, pathRelative } from '../../../../util/path'
import { CatalogComponentIcon } from '../../components/ComponentIcon'

import styles from './SourceSetDescendentComponents.module.scss'

export const SOURCE_SET_DESCENDENT_COMPONENTS_FRAGMENT = gql`
    fragment SourceSetDescendentComponentsFields on SourceLocationSet {
        descendentComponents {
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
    }
`

type SourceSetDescendentComponent = SourceSetDescendentComponentsFields['descendentComponents'][number]

interface Props extends FileSpec {
    repoID: Scalars['ID']
    descendentComponents: SourceSetDescendentComponentsFields['descendentComponents']
    className?: string
}

export const SourceSetDescendentComponents: React.FunctionComponent<Props> = ({
    descendentComponents,
    repoID,
    filePath,
    className,
}) => (
    <section className={className}>
        {descendentComponents && descendentComponents.length > 0 && (
            <ul className={classNames('list-unstyled', styles.boxGrid)}>
                {descendentComponents.map(component => (
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

const ComponentGridItem: React.FunctionComponent<{
    component: SourceSetDescendentComponent
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
    sourceLocation: SourceSetDescendentComponent['sourceLocations'][number]
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
