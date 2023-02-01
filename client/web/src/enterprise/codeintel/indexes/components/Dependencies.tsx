import { FunctionComponent, useCallback } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiMapSearch } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'
import { Observable } from 'rxjs'

import { H3, Icon, Link, Text, Tooltip } from '@sourcegraph/wildcard'

import {
    Connection,
    FilteredConnection,
    FilteredConnectionQueryArguments,
} from '../../../../components/FilteredConnection'
import { PreciseIndexFields } from '../../../../graphql-operations'
import { queryDependencyGraph as defaultQueryDependencyGraph } from '../hooks/queryDependencyGraph'

import { PreciseIndexLastUpdated } from './CodeIntelLastUpdated'
import { ProjectDescription } from './ProjectDescription'

import styles from './Dependencies.module.scss'

export interface DependencyListProps {
    index: PreciseIndexFields
    history: H.History
    queryDependencyGraph?: typeof defaultQueryDependencyGraph
}

export const DependenciesList: FunctionComponent<DependencyListProps> = ({
    index,
    history,
    queryDependencyGraph = defaultQueryDependencyGraph,
}) => {
    const apolloClient = useApolloClient()
    const queryDependencies = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryDependencyGraph({ ...args, dependencyOf: index.id }, apolloClient),
        [index, queryDependencyGraph, apolloClient]
    )

    return (
        <DependencyOrDependentsPanel
            noun="dependency"
            pluralNoun="dependencies"
            queryConnection={queryDependencies}
            history={history}
        />
    )
}

export interface DependentListProps {
    index: PreciseIndexFields
    history: H.History
    queryDependencyGraph?: typeof defaultQueryDependencyGraph
}

export const DependentsList: FunctionComponent<DependentListProps> = ({
    index,
    history,
    queryDependencyGraph = defaultQueryDependencyGraph,
}) => {
    const apolloClient = useApolloClient()
    const queryDependents = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryDependencyGraph({ ...args, dependentOf: index.id }, apolloClient),
        [index, queryDependencyGraph, apolloClient]
    )

    return (
        <DependencyOrDependentsPanel
            noun="dependent"
            pluralNoun="dependents"
            queryConnection={queryDependents}
            history={history}
        />
    )
}

interface DependencyOrDependentsPanelProps {
    noun: string
    pluralNoun: string
    queryConnection: (args: FilteredConnectionQueryArguments) => Observable<Connection<PreciseIndexFields>>
    history: H.History
}

const DependencyOrDependentsPanel: FunctionComponent<DependencyOrDependentsPanelProps> = ({
    noun,
    pluralNoun,
    queryConnection,
    history,
}) => (
    <FilteredConnection
        listComponent="div"
        listClassName={classNames(styles.grid, 'mb-3')}
        noun={noun}
        pluralNoun={pluralNoun}
        nodeComponent={DependencyOrDependentNode}
        nodeComponentProps={{ history }}
        queryConnection={queryConnection}
        cursorPaging={true}
        useURLQuery={false}
        emptyElement={<EmptyDependencyOrDependents pluralNoun={pluralNoun} />}
    />
)

interface DependencyOrDependentNodeProps {
    node: PreciseIndexFields
    history: H.History
}

const DependencyOrDependentNode: FunctionComponent<DependencyOrDependentNodeProps> = ({ node, history }) => (
    <div
        className={classNames(styles.grid, 'px-4')}
        onClick={() => {
            if (node.projectRoot) {
                history.push(`/${node.projectRoot.repository.name}/-/code-graph/indexes/${node.id}`)
            }
        }}
        aria-hidden={true}
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
