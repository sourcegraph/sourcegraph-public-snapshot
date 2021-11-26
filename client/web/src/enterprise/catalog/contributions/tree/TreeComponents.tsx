import classNames from 'classnames'
import FolderIcon from 'mdi-react/FolderIcon'
import React from 'react'
import { Link } from 'react-router-dom'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { FileSpec } from '@sourcegraph/shared/src/util/url'

import {
    DescendentComponentForTreeEntryFields,
    ExactComponentForTreeEntryFields,
    ComponentsForTreeEntryResult,
    ComponentsForTreeEntryVariables,
} from '../../../../graphql-operations'
import { pathHasPrefix, pathRelative } from '../../../../util/path'
import { ComponentIcon } from '../../components/ComponentIcon'
import { COMPONENT_OWNER_FRAGMENT } from '../../components/entity-owner/gql'
import {
    COMPONENT_STATUS_FRAGMENT,
    COMPONENT_CODE_OWNERS_FRAGMENT,
    COMPONENT_AUTHORS_FRAGMENT,
    COMPONENT_USAGE_PEOPLE_FRAGMENT,
} from '../../pages/component/gql'
import { OverviewStatusContexts } from '../../pages/component/OverviewStatusContexts'

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
        exactComponents: components(path: $path, recursive: false) {
            ...ExactComponentForTreeEntryFields
        }
        descendentComponents: components(path: $path, recursive: true) {
            ...DescendentComponentForTreeEntryFields
        }
    }

    fragment ExactComponentForTreeEntryFields on Component {
        __typename
        id
        name
        kind
        description
        lifecycle
        url
        ...ComponentOwnerFields
        ...ComponentStatusFields
        ...ComponentCodeOwnersFields
        ...ComponentAuthorsFields
        ...ComponentUsagePeopleFields
    }

    fragment DescendentComponentForTreeEntryFields on Component {
        __typename
        id
        name
        kind
        description
        url
        sourceLocations {
            ... on GitTree {
                repository {
                    id
                }
            }
            path
            url
        }
    }

    ${COMPONENT_OWNER_FRAGMENT}
    ${COMPONENT_STATUS_FRAGMENT}
    ${COMPONENT_CODE_OWNERS_FRAGMENT}
    ${COMPONENT_AUTHORS_FRAGMENT}
    ${COMPONENT_USAGE_PEOPLE_FRAGMENT}
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

    const exactComponents = (data && data.node?.__typename === 'Repository' && data.node.exactComponents) || null
    const descendentComponents =
        (data &&
            data.node?.__typename === 'Repository' &&
            data.node.descendentComponents.filter(component => component.id !== exactComponents?.[0]?.id)) ||
        null

    if (
        (!descendentComponents || descendentComponents.length === 0) &&
        (!exactComponents || exactComponents.length === 0)
    ) {
        return null
    }

    return (
        <section className={className}>
            {exactComponents && exactComponents.length > 0 && (
                <ComponentDetail
                    component={exactComponents[0]}
                    className={classNames('px-3 pt-3 border border-secondary rounded', styles.componentDetail)}
                />
            )}
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
}

const ComponentDetail: React.FunctionComponent<{
    component: ExactComponentForTreeEntryFields
    className?: string
}> = ({ component, className }) => (
    <div className={className}>
        <h3>
            <Link to={component.url} className="d-flex align-items-center font-weight-bold">
                <ComponentIcon component={component} className="icon-inline mr-1 text-muted" /> {component.name}
            </Link>
        </h3>
        <div className="text-muted small mb-2">
            {component.__typename === 'Component' && `${component.kind[0]}${component.kind.slice(1).toLowerCase()}`}
        </div>
        {component.description && <p>{component.description}</p>}
        <OverviewStatusContexts component={component} itemClassName="mb-3" />
    </div>
)

const ComponentGridItem: React.FunctionComponent<{
    component: DescendentComponentForTreeEntryFields
    treeRepoID: Scalars['ID']
    treePath: string
    tag?: 'li'
    className?: string
    linkBigClickAreaClassName?: string
}> = ({ component, treeRepoID, treePath, tag: Tag = 'li', className, linkBigClickAreaClassName }) => {
    const relevantSourceLocation = component.sourceLocations.find(
        sourceLocation =>
            sourceLocation.__typename === 'GitTree' &&
            sourceLocation.repository.id === treeRepoID &&
            pathHasPrefix(sourceLocation.path, treePath)
    )
    if (!relevantSourceLocation) {
        throw new Error('unable to determine relevant source location')
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
                        <ComponentIcon component={component} className="icon-inline mr-1 text-muted" />
                        {component.name}
                    </Link>
                </h4>
                {component.description && (
                    <p className={classNames('my-1 small', styles.boxGridItemBody)}>{component.description}</p>
                )}
            </div>
            <div className="small mt-1">
                <Link
                    to={
                        relevantSourceLocation.url /* TODO(sqs): this takes you away from the current rev back to HEAD */
                    }
                    className={classNames('d-flex align-items-center text-muted', linkBigClickAreaClassName)}
                >
                    <FolderIcon className="icon-inline mr-1 flex-shrink-0" />{' '}
                    <span className="text-truncate">{pathRelative(treePath, relevantSourceLocation.path)}</span>
                </Link>
            </div>
        </Tag>
    )
}
