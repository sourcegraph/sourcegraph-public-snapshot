import { FunctionComponent, useCallback } from 'react'

import { useApolloClient } from '@apollo/client'
import classNames from 'classnames'
import * as H from 'history'

import { isErrorLike } from '@sourcegraph/common'
import { H3, Link, Tooltip } from '@sourcegraph/wildcard'

import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import { PreciseIndexFields } from '../../../../graphql-operations'
import { PreciseIndexLastUpdated } from '../components/CodeIntelLastUpdated'
import { ProjectDescription } from './ProjectDescription'
import { queryDependencyGraph as defaultQueryDependencyGraph } from '../hooks/queryDependencyGraph'

import styles from './Dependencies.module.scss'

export interface DependencyListProps {
    index: PreciseIndexFields
    history: H.History
    location: H.Location
    queryDependencyGraph?: typeof defaultQueryDependencyGraph
}

export const DependenciesPanel: FunctionComponent<DependencyListProps> = ({
    index,
    history,
    location,
    queryDependencyGraph = defaultQueryDependencyGraph,
}) => {
    const apolloClient = useApolloClient()
    const queryDependencies = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            if (index && !isErrorLike(index)) {
                return queryDependencyGraph({ ...args, dependencyOf: index.id }, apolloClient)
            }
            throw new Error('unreachable: queryDependencies referenced with invalid upload')
        },
        [index, queryDependencyGraph, apolloClient]
    )

    return (
        <FilteredConnection
            listComponent="div"
            listClassName={classNames(styles.grid, 'mb-3')}
            noun="dependency"
            pluralNoun="dependencies"
            nodeComponent={DependencyOrDependentNode}
            nodeComponentProps={{ history }}
            queryConnection={queryDependencies}
            history={history}
            location={location}
            cursorPaging={true}
            useURLQuery={false}
            // emptyElement={<EmptyDependencies />}
        />
    )
}

export interface DependentListProps {
    index: PreciseIndexFields
    history: H.History
    location: H.Location
    queryDependencyGraph?: typeof defaultQueryDependencyGraph
}

export const DependentsPanel: FunctionComponent<DependentListProps> = ({
    index,
    history,
    location,
    queryDependencyGraph = defaultQueryDependencyGraph,
}) => {
    const apolloClient = useApolloClient()
    const queryDependents = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            if (index && !isErrorLike(index)) {
                return queryDependencyGraph({ ...args, dependentOf: index.id }, apolloClient)
            }

            throw new Error('unreachable: queryDependents referenced with invalid upload')
        },
        [index, queryDependencyGraph, apolloClient]
    )

    return (
        <FilteredConnection
            listComponent="div"
            listClassName={classNames(styles.grid, 'mb-3')}
            noun="dependent"
            pluralNoun="dependents"
            nodeComponent={DependencyOrDependentNode}
            nodeComponentProps={{ history }}
            queryConnection={queryDependents}
            history={history}
            location={location}
            cursorPaging={true}
            useURLQuery={false}
            // emptyElement={<EmptyDependents />}
        />
    )
}

interface DependencyOrDependentNodeProps {
    node: PreciseIndexFields
    history: H.History
}

const DependencyOrDependentNode: FunctionComponent<React.PropsWithChildren<DependencyOrDependentNodeProps>> = ({
    node,
    history,
}) => (
    <div
        className={classNames(styles.grid, 'px-4')}
        onClick={() => {
            if (node.projectRoot) {
                history.push(`/${node.projectRoot.repository.name}/-/code-graph/indexes/${node.id}`)
            }
        }}
    >
        <div>
            <H3 className="m-0 mb-1">
                {node.projectRoot ? (
                    <Link to={node.projectRoot.repository.url} onClick={event => event.stopPropagation()}>
                        {node.projectRoot.repository.name}
                    </Link>
                ) : (
                    <span>Unknown repository</span>
                )}
            </H3>
        </div>

        <div>
            <span className="mr-2 d-block d-mdinline-block">
                <ProjectDescription index={node} onLinkClick={event => event.stopPropagation()} />
            </span>

            <small className="text-mute">
                <PreciseIndexLastUpdated index={node} />{' '}
                {node.shouldReindex && (
                    <Tooltip content="This index has been marked as replaceable by auto-indexing.">
                        <span className={classNames(styles.tag, 'ml-1 rounded')}>(replaceable by auto-indexing)</span>
                    </Tooltip>
                )}
            </small>
        </div>
    </div>
)
