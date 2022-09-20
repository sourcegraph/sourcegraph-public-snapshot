import React, { useCallback, useEffect, useRef, useState } from 'react'

import { lowerCase, upperFirst } from 'lodash'
import { useHistory } from 'react-router'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { Input, Select } from '@sourcegraph/wildcard'

import { BatchSpecWorkspaceState } from '../../../../../graphql-operations'
import { isValidBatchSpecWorkspaceState } from '../util'

/** We exclude pending as a filter option, because it's not a valid state on the execution page. */
const STATES_WITHOUT_PENDING = Object.values(BatchSpecWorkspaceState).filter(
    value => value !== BatchSpecWorkspaceState.PENDING
)

export interface WorkspaceFilters {
    state: BatchSpecWorkspaceState | null
    search: string | null
}

interface WorkspaceFilterRowProps {
    onFiltersChange: (newFilters: WorkspaceFilters) => void
}

export const WorkspaceFilterRow: React.FunctionComponent<React.PropsWithChildren<WorkspaceFilterRowProps>> = ({
    onFiltersChange,
}) => {
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
        <div className="d-flex align-items-center mb-2">
            <Form className="d-flex flex-grow-1 mr-2" onSubmit={onSubmit}>
                <Input
                    className="m-0 flex-1"
                    type="search"
                    ref={searchElement}
                    defaultValue={search}
                    placeholder="Search repository name"
                    aria-label="Search repository name"
                />
            </Form>
            <WorkspaceFilter<BatchSpecWorkspaceState>
                values={STATES_WITHOUT_PENDING}
                label="State"
                selected={state}
                onChange={setState}
                className="m-0"
            />
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
            isCustomStyle={true}
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
