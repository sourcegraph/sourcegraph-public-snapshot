import React, { type FC, useCallback, useContext, useEffect, useRef } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'

import { Input, Form } from '@sourcegraph/wildcard'

import { ChangesetSpecOperation, ChangesetState } from '../../../../graphql-operations'
import { ChangesetFilter } from '../../ChangesetFilter'
import { BatchChangePreviewContext } from '../BatchChangePreviewContext'

export interface PreviewFilters {
    search: string | null
    currentState: ChangesetState | null
    action: ChangesetSpecOperation | null
}

export interface PreviewFilterRowProps {}

export const PreviewFilterRow: FC<PreviewFilterRowProps> = props => {
    const location = useLocation()
    const navigate = useNavigate()

    const searchElement = useRef<HTMLInputElement | null>(null)

    // `BatchChangePreviewContext` is responsible for managing the filter arguments for
    // the `applyPreview` connection query.
    const { filters, setFilters, setFiltersChanged } = useContext(BatchChangePreviewContext)

    const onSubmit = useCallback(
        (event: React.FormEvent<HTMLFormElement>): void => {
            event.preventDefault()
            setFilters({ ...filters, search: searchElement.current?.value || null })
            if (filters.search !== searchElement.current?.value) {
                setFiltersChanged(true)
            }
        },
        [setFilters, filters, setFiltersChanged]
    )

    const setAction = useCallback(
        (action: ChangesetSpecOperation | undefined) => {
            setFilters({ ...filters, action: action || null })
        },
        [filters, setFilters]
    )

    const setCurrentState = useCallback(
        (currentState: ChangesetState | undefined) => {
            setFilters({ ...filters, currentState: currentState || null })
        },
        [filters, setFilters]
    )

    useEffect(() => {
        const urlParameters = new URLSearchParams(location.search)
        const { search, action, currentState } = filters

        if (search) {
            urlParameters.set('search', search)
        } else {
            urlParameters.delete('search')
        }

        if (action) {
            urlParameters.set('action', action)
        } else {
            urlParameters.delete('action')
        }

        if (currentState) {
            urlParameters.set('current_state', currentState)
        } else {
            urlParameters.delete('current_state')
        }

        if (location.search !== urlParameters.toString()) {
            navigate({ search: urlParameters.toString() })
        }

        // We cannot depend on location.search, since it's modified by this hook and that would cause an infinite render loop.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [filters, navigate])

    return (
        <div className="row no-gutters">
            <div className="m-0 col">
                <Form className="form-inline d-flex mb-2" onSubmit={onSubmit}>
                    <Input
                        className="flex-grow-1"
                        inputClassName="flex-grow-1"
                        type="search"
                        ref={searchElement}
                        defaultValue={filters.search ?? undefined}
                        placeholder="Search title and repository name"
                        aria-label="Search title and repository name"
                    />
                </Form>
            </div>
            <div className="w-100 d-block d-md-none" />
            <div className="m-0 col col-md-auto">
                <div className="row no-gutters">
                    <div className="col mb-2 ml-0 ml-md-2">
                        <ChangesetFilter<ChangesetState>
                            values={Object.values(ChangesetState)}
                            label="Current state"
                            selected={filters.currentState ?? undefined}
                            onChange={setCurrentState}
                            className="w-100"
                        />
                    </div>
                    <div className="col mb-2 ml-2">
                        <ChangesetFilter<ChangesetSpecOperation>
                            values={Object.values(ChangesetSpecOperation)}
                            label="Actions"
                            selected={filters.action ?? undefined}
                            onChange={setAction}
                            className="w-100"
                        />
                    </div>
                </div>
            </div>
        </div>
    )
}
