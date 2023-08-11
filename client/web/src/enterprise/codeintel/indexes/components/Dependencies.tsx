import { type FunctionComponent, useCallback } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiChevronRight, mdiMapSearch } from '@mdi/js'
import classNames from 'classnames'
import type { Observable } from 'rxjs'

import { H3, Icon, Link, Text, Tooltip } from '@sourcegraph/wildcard'

import {
    type Connection,
    FilteredConnection,
    type FilteredConnectionQueryArguments,
} from '../../../../components/FilteredConnection'
import type { PreciseIndexFields } from '../../../../graphql-operations'
import { queryDependencyGraph as defaultQueryDependencyGraph } from '../hooks/queryDependencyGraph'

import { PreciseIndexLastUpdated } from './CodeIntelLastUpdated'
import { ProjectDescription } from './ProjectDescription'

import styles from './Dependencies.module.scss'

export interface DependencyListProps {
    index: PreciseIndexFields
    queryDependencyGraph?: typeof defaultQueryDependencyGraph
}

export const DependenciesList: FunctionComponent<DependencyListProps> = ({
    index,
    queryDependencyGraph = defaultQueryDependencyGraph,
}) => {
    const apolloClient = useApolloClient()
    const queryDependencies = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryDependencyGraph({ ...args, dependencyOf: index.id }, apolloClient),
        [index, queryDependencyGraph, apolloClient]
    )

    return (
        <DependencyOrDependentsPanel noun="dependency" pluralNoun="dependencies" queryConnection={queryDependencies} />
    )
}

export interface DependentListProps {
    index: PreciseIndexFields
    queryDependencyGraph?: typeof defaultQueryDependencyGraph
}

export const DependentsList: FunctionComponent<DependentListProps> = ({
    index,
    queryDependencyGraph = defaultQueryDependencyGraph,
}) => {
    const apolloClient = useApolloClient()
    const queryDependents = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryDependencyGraph({ ...args, dependentOf: index.id }, apolloClient),
        [index, queryDependencyGraph, apolloClient]
    )

    return <DependencyOrDependentsPanel noun="dependent" pluralNoun="dependents" queryConnection={queryDependents} />
}

interface DependencyOrDependentsPanelProps {
    noun: string
    pluralNoun: string
    queryConnection: (args: FilteredConnectionQueryArguments) => Observable<Connection<PreciseIndexFields>>
}

const DependencyOrDependentsPanel: FunctionComponent<DependencyOrDependentsPanelProps> = ({
    noun,
    pluralNoun,
    queryConnection,
}) => (
    <FilteredConnection
        listComponent="div"
        listClassName={classNames(styles.list, 'mb-3')}
        noun={noun}
        pluralNoun={pluralNoun}
        nodeComponent={DependencyOrDependentNode}
        queryConnection={queryConnection}
        cursorPaging={true}
        useURLQuery={false}
        emptyElement={<EmptyDependencyOrDependents pluralNoun={pluralNoun} />}
    />
)

interface DependencyOrDependentNodeProps {
    node: PreciseIndexFields
}

const DependencyOrDependentNode: FunctionComponent<DependencyOrDependentNodeProps> = ({ node }) => (
    <div className={classNames(styles.listItem, 'px-4')}>
        <div>
            <div>
                <H3 className="m-0 mb-1">
                    {node.projectRoot ? (
                        <Link to={node.projectRoot.repository.url}>{node.projectRoot.repository.name}</Link>
                    ) : (
                        <span>Unknown repository</span>
                    )}
                </H3>
            </div>

            <div>
                <span className="mr-2 d-block d-mdinline-block">
                    <ProjectDescription index={node} />
                </span>

                <small className="text-mute">
                    <PreciseIndexLastUpdated index={node} />{' '}
                    {node.shouldReindex && (
                        <Tooltip content="This index has been marked as replaceable by auto-indexing.">
                            <span className={classNames(styles.tag, 'ml-1 rounded')}>
                                (replaceable by auto-indexing)
                            </span>
                        </Tooltip>
                    )}
                </small>
            </div>
        </div>
        {node.projectRoot && (
            <Link
                to={`/${node.projectRoot.repository.name}/-/code-graph/indexes/${node.id}`}
                className="d-flex justify-content-end align-items-center align-self-stretch p-0"
            >
                <Icon svgPath={mdiChevronRight} inline={false} aria-label="View details" />
            </Link>
        )}
    </div>
)

interface EmptyDependencyOrDependentsProps {
    pluralNoun: string
}

const EmptyDependencyOrDependents: React.FunctionComponent<EmptyDependencyOrDependentsProps> = ({ pluralNoun }) => (
    <Text alignment="center" className="text-muted w-100 mb-0 mt-1">
        <Icon className="mb-2" svgPath={mdiMapSearch} inline={false} aria-hidden={true} />
        <br />
        No {pluralNoun}.
    </Text>
)
