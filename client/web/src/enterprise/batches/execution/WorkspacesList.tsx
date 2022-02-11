import classNames from 'classnames'
import { lowerCase, upperFirst } from 'lodash'
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { useHistory } from 'react-router'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { Input, Link, Select } from '@sourcegraph/wildcard'

import { DiffStat } from '../../../components/diff/DiffStat'
import {
    ConnectionContainer,
    ConnectionError,
    ConnectionList,
    ConnectionLoading,
    ConnectionSummary,
    ShowMoreButton,
    SummaryContainer,
} from '../../../components/FilteredConnection/ui'
import { BatchSpecWorkspaceListFields, BatchSpecWorkspaceState, Scalars } from '../../../graphql-operations'
import { Branch } from '../Branch'

import { useWorkspacesListConnection } from './backend'
import { isValidBatchSpecWorkspaceState } from './util'
import styles from './WorkspacesList.module.scss'
import { WorkspaceStateIcon } from './WorkspaceStateIcon'

export interface WorkspacesListProps {
    batchSpecID: Scalars['ID']
    /** The currently selected workspace node id. Will be highlighted. */
    selectedNode?: Scalars['ID']
}

export const WorkspacesList: React.FunctionComponent<WorkspacesListProps> = ({ batchSpecID, selectedNode }) => {
    const [filters, setFilters] = useState<WorkspaceFilters>()
    const { loading, hasNextPage, fetchMore, connection, error } = useWorkspacesListConnection(
        batchSpecID,
        filters?.search ?? null,
        filters?.state ?? null
    )

    return (
        <ConnectionContainer>
            {error && <ConnectionError errors={[error.message]} />}
            <WorkspaceFilterRow onFiltersChange={setFilters} />
            <ConnectionList as="ul" className="list-group list-group-flush">
                {connection?.nodes?.map(node => (
                    <WorkspaceNode key={node.id} node={node} selectedNode={selectedNode} />
                ))}
            </ConnectionList>
            {/* We don't want to flash a loader on reloads: */}
            {loading && !connection && <ConnectionLoading />}
            {connection && (
                <SummaryContainer centered={true}>
                    <ConnectionSummary
                        noSummaryIfAllNodesVisible={true}
                        first={20}
                        connection={connection}
                        noun="workspace"
                        pluralNoun="workspaces"
                        hasNextPage={hasNextPage}
                    />
                    {hasNextPage && <ShowMoreButton centered={true} onClick={fetchMore} />}
                </SummaryContainer>
            )}
        </ConnectionContainer>
    )
}

interface WorkspaceNodeProps {
    node: BatchSpecWorkspaceListFields
    selectedNode?: Scalars['ID']
}

const WorkspaceNode: React.FunctionComponent<WorkspaceNodeProps> = ({ node, selectedNode }) => (
    <li className={classNames('list-group-item', node.id === selectedNode && styles.workspaceSelected)}>
        <Link to={`?workspace=${node.id}`}>
            <div className={classNames(styles.workspaceRepo, 'd-flex justify-content-between mb-1')}>
                <span>
                    <WorkspaceStateIcon
                        cachedResultFound={node.cachedResultFound}
                        state={node.state}
                        className={classNames(styles.workspaceListIcon, 'mr-2 flex-shrink-0')}
                    />
                </span>
                <strong className={classNames(styles.workspaceName, 'flex-grow-1')}>{node.repository.name}</strong>
                {node.diffStat && <DiffStat {...node.diffStat} expandedCounts={true} />}
            </div>
            <Branch name={node.branch.abbrevName} />
        </Link>
    </li>
)

export interface WorkspaceFilters {
    state: BatchSpecWorkspaceState | null
    search: string | null
}

export interface WorkspaceFilterRowProps {
    onFiltersChange: (newFilters: WorkspaceFilters) => void
}

export const WorkspaceFilterRow: React.FunctionComponent<WorkspaceFilterRowProps> = ({ onFiltersChange }) => {
    const history = useHistory()
    const searchElement = useRef<HTMLInputElement | null>(null)
    const [state, setState] = useState<BatchSpecWorkspaceState | undefined>(() => {
        const searchParameters = new URLSearchParams(history.location.search)
        const value = searchParameters.get('state')
        return value && isValidBatchSpecWorkspaceState(value) ? value : undefined
    })
    const [search, setSearch] = useState<string | undefined>(() => {
        const searchParameters = new URLSearchParams(history.location.search)
        return searchParameters.get('search') ?? undefined
    })
    useEffect(() => {
        const searchParameters = new URLSearchParams(history.location.search)
        if (state) {
            searchParameters.set('state', state)
        } else {
            searchParameters.delete('state')
        }
        if (search) {
            searchParameters.set('search', search)
        } else {
            searchParameters.delete('search')
        }
        if (history.location.search !== searchParameters.toString()) {
            history.replace({ ...history.location, search: searchParameters.toString() })
        }
        // Update the filters in the parent component.
        onFiltersChange({
            state: state || null,
            search: search || null,
        })
        // We cannot depend on the history, since it's modified by this hook and that would cause an infinite render loop.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [state, search])

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(event => {
        event?.preventDefault()
        setSearch(searchElement.current?.value)
    }, [])

    return (
        <div className="row no-gutters mr-1">
            <div className="m-0 col">
                <Form className="d-flex mb-2" onSubmit={onSubmit}>
                    <Input
                        className="flex-grow-1"
                        type="search"
                        ref={searchElement}
                        defaultValue={search}
                        placeholder="Search repository name"
                    />
                </Form>
            </div>
            <div className="w-100 d-block d-md-none" />
            <div className="m-0 col col-md-auto">
                <div className="row no-gutters">
                    <div className="col mb-2 ml-0 ml-md-2">
                        <WorkspaceFilter<BatchSpecWorkspaceState>
                            values={Object.values(BatchSpecWorkspaceState)}
                            label="State"
                            selected={state}
                            onChange={setState}
                            className="w-100"
                        />
                    </div>
                </div>
            </div>
        </div>
    )
}

export interface WorkspaceFilterProps<T extends string> {
    label: string
    values: T[]
    selected: T | undefined
    onChange: (value: T | undefined) => void
    className?: string
}

export const WorkspaceFilter = <T extends string>({
    label,
    values,
    selected,
    onChange,
    className,
}: WorkspaceFilterProps<T>): React.ReactElement<WorkspaceFilterProps<T>> => {
    const innerOnChange = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(
        event => {
            onChange((event.target.value ?? undefined) as T | undefined)
        },
        [onChange]
    )

    return (
        <Select
            id="workspace-state"
            className={className}
            value={selected}
            onChange={innerOnChange}
            aria-label="Filter by workspace state"
        >
            <option value="">{label}</option>
            {values.map(state => (
                <option value={state} key={state}>
                    {upperFirst(lowerCase(state))}
                </option>
            ))}
        </Select>
    )
}
