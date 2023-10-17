import React, { type FC, useCallback, useRef, useState } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'

import { Input, Form } from '@sourcegraph/wildcard'

import type { WorkspacePreviewFilters } from './useWorkspaces'

export interface WorkspacePreviewFilterRowProps {
    /** Whether or not the filter form should be disabled. */
    disabled: boolean
    /** Method to invoke to capture a change in the active filters applied. */
    onFiltersChange: (newFilters: WorkspacePreviewFilters) => void
}

export const WorkspacePreviewFilterRow: FC<WorkspacePreviewFilterRowProps> = ({ disabled, onFiltersChange }) => {
    const navigate = useNavigate()
    const location = useLocation()

    const searchElement = useRef<HTMLInputElement | null>(null)
    const [search, setSearch] = useState<string | undefined>(() => {
        const searchParameters = new URLSearchParams(location.search)
        return searchParameters.get('search') ?? undefined
    })

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        event => {
            event?.preventDefault()
            const value = searchElement.current?.value
            setSearch(value)

            // Update the location, too.
            const searchParameters = new URLSearchParams(location.search)
            if (value) {
                searchParameters.set('search', value)
            } else {
                searchParameters.delete('search')
            }
            if (location.search !== searchParameters.toString()) {
                navigate({ search: searchParameters.toString() }, { replace: true })
            }
            // Update the filters in the parent component.
            onFiltersChange({
                search: value || null,
            })
        },
        [navigate, location.search, onFiltersChange]
    )

    return (
        <div className="w-100 row">
            <div className="m-0 p-0 col">
                <Form className="d-flex mb-2" onSubmit={onSubmit}>
                    <Input
                        disabled={disabled}
                        className="flex-grow-1"
                        type="search"
                        ref={searchElement}
                        defaultValue={search}
                        placeholder="Search repository name"
                        aria-label="Search repository name"
                    />
                </Form>
            </div>
        </div>
    )
}
