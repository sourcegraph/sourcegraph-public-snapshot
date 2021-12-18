import classNames from 'classnames'
import React, { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'

import { dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'

import { useConnection } from '../../../../components/FilteredConnection/hooks/useConnection'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    SummaryContainer,
    ConnectionSummary,
    ShowMoreButton,
} from '../../../../components/FilteredConnection/ui'
import {
    CatalogHealthResult,
    CatalogHealthVariables,
    ComponentHealthFields,
    ComponentStatusFields,
    ComponentStatusState,
} from '../../../../graphql-operations'
import { ComponentFiltersProps } from '../../core/entity-filters'
import { ComponentStateIndicator } from '../../pages/overview/components/entity-state-indicator/EntityStateIndicator'
import { ComponentIcon } from '../ComponentIcon'
import { EntityOwner } from '../entity-owner/EntityOwner'

import styles from './CatalogHealthTable.module.scss'
import { CATALOG_HEALTH } from './gql'

interface Props extends Pick<ComponentFiltersProps, 'filters'> {
    queryScope?: string
    className?: string
}

const FIRST = 20

export const CatalogHealthTable: React.FunctionComponent<Props> = ({ filters, queryScope, className }) => {
    const { connection, error, loading, fetchMore, hasNextPage } = useConnection<
        CatalogHealthResult,
        CatalogHealthVariables,
        ComponentHealthFields
    >({
        query: CATALOG_HEALTH,
        variables: {
            query: `${queryScope || ''} ${filters.query || ''}`,
            first: FIRST,
            after: null,
        },
        options: {
            useURL: true,
            fetchPolicy: 'cache-and-network',
        },
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            return data.catalog.entities
        },
    })

    const [useColor, setUseColor] = useState(true)

    return (
        <>
            <ConnectionContainer className={classNames('position-relative', className)}>
                <button
                    type="button"
                    style={{ position: 'absolute', top: '-6px', left: '113px', width: '60px' }}
                    className="btn btn-sm btn-link p-0 text-muted"
                    onClick={() => setUseColor(previous => !previous)}
                >
                    {useColor ? 'Red/green' : 'Text'}
                </button>
                {error && <ConnectionError errors={[error.message]} />}
                {connection && connection.nodes.length > 0 && (
                    <div className="table-responsive">
                        <ConnectionList className={classNames('table border-bottom', styles.table)} as="table">
                            <CatalogHealthTableContent nodes={connection.nodes} useColor={useColor} />
                        </ConnectionList>
                    </div>
                )}
                {loading && <ConnectionLoading className="py-2" />}
                {connection && (
                    <SummaryContainer centered={true}>
                        <ConnectionSummary
                            noSummaryIfAllNodesVisible={true}
                            first={FIRST}
                            connection={connection}
                            noun="entity"
                            pluralNoun="entities"
                            hasNextPage={hasNextPage}
                            emptyElement={<p>No results found</p>}
                        />
                        {hasNextPage && <ShowMoreButton onClick={fetchMore} />}
                    </SummaryContainer>
                )}
            </ConnectionContainer>
        </>
    )
}

type StatusContextNameAndTitle = Pick<ComponentStatusFields['status']['contexts'][0], 'name' | 'title'>

const CatalogHealthTableContent: React.FunctionComponent<{
    nodes: ComponentHealthFields[]
    useColor: boolean
}> = ({ nodes, useColor }) => {
    const statusContextNames = useMemo<StatusContextNameAndTitle[]>(() => {
        const nameTitle = new Map<string, string>()
        for (const node of nodes) {
            for (const statusContext of node.status.contexts) {
                nameTitle.set(statusContext.name, statusContext.title)
            }
        }
        return [...nameTitle.entries()]
            .map(([name, title]) => ({ name, title }))
            .sort((a, b) => a.name.localeCompare(b.name))
    }, [nodes])

    const TH_CLASS_NAME = 'text-muted small font-weight-normal py-2'

    return (
        <>
            <colgroup>
                <col className={styles.colName} />
                <col className={styles.colOwner} />
                <col className={styles.colStatusContext} span={1 + statusContextNames.length} />
            </colgroup>
            <thead>
                <tr>
                    <th className={classNames(TH_CLASS_NAME, styles.headerEntityName)} scope="col">
                        Name
                    </th>
                    <th className={classNames(TH_CLASS_NAME, styles.headerEntityOwner)} scope="col">
                        Owner
                    </th>
                    <th className={classNames(TH_CLASS_NAME, styles.headerCombinedStatus)} scope="col">
                        Overall
                    </th>
                    {statusContextNames.map(({ name, title }) => (
                        <th
                            key={name}
                            className={classNames(TH_CLASS_NAME, 'text-truncate', styles.headerStatusContext)}
                            scope="col"
                            title={title ? name : undefined}
                        >
                            {title || name}
                        </th>
                    ))}
                </tr>
            </thead>

            <tbody>
                {nodes.map(node => (
                    <CatalogHealthTableRow
                        key={node.id}
                        node={node}
                        statusContextNames={statusContextNames}
                        useColor={useColor}
                    />
                ))}
            </tbody>
        </>
    )
}

const CatalogHealthTableRow: React.FunctionComponent<{
    node: ComponentHealthFields
    statusContextNames: StatusContextNameAndTitle[]
    useColor: boolean
}> = ({ node, statusContextNames, useColor }) => {
    const score =
        node.status.contexts.length > 0
            ? node.status.contexts.filter(
                  ({ state }) => state === ComponentStatusState.SUCCESS || state === ComponentStatusState.INFO
              ).length / node.status.contexts.length
            : 0
    return (
        <tr>
            <td>
                <h3 className={classNames('h6 font-weight-bold mb-0 d-flex align-items-center')}>
                    <Link to={node.url} className={classNames('d-block text-truncate')}>
                        <ComponentIcon
                            entity={node}
                            className={classNames('icon-inline mr-1 flex-shrink-0 text-muted')}
                        />
                        {node.name}
                    </Link>
                    <ComponentStateIndicator entity={node} className="ml-1" />
                </h3>
            </td>
            <td className={styles.cellEntityOwner}>
                <EntityOwner owner={node.owner} className="text-nowrap d-flex" blankIfNone={true} />
            </td>
            <ComponentStatusStateCell
                state={node.status.state}
                targetURL={node.url}
                description={`Combined status for ${node.name}: ${node.status.state.toLowerCase()}`}
            >
                {Math.round(100 * score)}%
            </ComponentStatusStateCell>
            {statusContextNames.map(({ name: statusContextName }) => {
                const status = node.status.contexts.find(statusContext => statusContext.name === statusContextName)
                return (
                    <ComponentStatusStateCell
                        key={statusContextName}
                        state={status ? status.state : null}
                        targetURL={status?.targetURL || node.url}
                        description={
                            status
                                ? `${status.name} status for ${node.name}: ${status.state.toLowerCase()}${
                                      status.description ? `\n${status.description}` : ''
                                  }`
                                : `No ${statusContextName} status for ${node.name}`
                        }
                    >
                        {status ? (
                            <small className={useColor ? 'sr-only' : 'font-weight-normal'}>
                                {status.state[0]}
                                {status.state.slice(1).toLowerCase()}
                            </small>
                        ) : null}
                    </ComponentStatusStateCell>
                )
            })}
        </tr>
    )
}

const ComponentStatusStateCell: React.FunctionComponent<{
    state: ComponentStatusState | null
    targetURL: string
    description?: string | null
}> = ({ state, targetURL, description, children }) => (
    <td
        className={classNames(
            'position-relative',
            styles.cellState,
            state ? CELL_CLASS_NAME_FOR_STATE[state] : styles.stateNull
        )}
        data-tooltip={description || undefined}
    >
        <Link to={targetURL} className="d-block stretched-link">
            <span className="sr-only">{state ? state.toLowerCase() : 'none'}</span>
            {children}
        </Link>
    </td>
)

const CELL_CLASS_NAME_FOR_STATE: Record<ComponentStatusState, string> = {
    EXPECTED: styles.stateExpected,
    ERROR: styles.stateError,
    FAILURE: styles.stateFailure,
    PENDING: styles.statePending,
    SUCCESS: styles.stateSuccess,
    INFO: styles.stateInfo,
}
