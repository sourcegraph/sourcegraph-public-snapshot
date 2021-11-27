import classNames from 'classnames'
import React, { ChangeEventHandler, useCallback, useRef, useState } from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { Resizable } from '@sourcegraph/shared/src/components/Resizable'
import { Button } from '@sourcegraph/wildcard'

import { Badge } from '../../../../../../components/Badge'
import { FeedbackPromptContent } from '../../../../../../nav/Feedback/FeedbackPrompt'
import { Popover } from '../../../../../insights/components/popover/Popover'
import { CatalogComponentFilters } from '../../../../core/component-filters'

import styles from './Sidebar.module.scss'

const SIZE_STORAGE_KEY = 'catalog-sidebar-size'

interface SidebarProps {
    filters: CatalogComponentFilters
    onFiltersChange: (newFilters: CatalogComponentFilters) => void
}

export const Sidebar: React.FunctionComponent<SidebarProps> = props => (
    <Resizable
        defaultSize={200}
        handlePosition="right"
        storageKey={SIZE_STORAGE_KEY}
        className={styles.resizable}
        element={<SidebarContent className="border-right w-100" {...props} />}
    />
)

const SidebarContent: React.FunctionComponent<SidebarProps & { className?: string }> = ({ className, ...props }) => (
    <div className={classNames('p-2 d-flex flex-column', className)}>
        <h2 className="h5 font-weight-bold">Catalog</h2>
        <ComponentFiltersForm {...props} />
        <div className="flex-1" />
        <FeedbackPopoverButton />
    </div>
)

const ComponentFiltersForm: React.FunctionComponent<SidebarProps> = ({ filters, onFiltersChange }) => {
    const onQueryChange = useCallback<ChangeEventHandler<HTMLInputElement>>(
        event => {
            onFiltersChange({ ...filters, query: event.target.value })
        },
        [filters, onFiltersChange]
    )
    const onOwnerChange = useCallback<ChangeEventHandler<HTMLInputElement>>(
        event => {
            onFiltersChange({ ...filters, owner: event.target.value })
        },
        [filters, onFiltersChange]
    )
    const onSystemChange = useCallback<ChangeEventHandler<HTMLInputElement>>(
        event => {
            onFiltersChange({ ...filters, system: event.target.value })
        },
        [filters, onFiltersChange]
    )
    const onTagsChange = useCallback<ChangeEventHandler<HTMLInputElement>>(
        event => {
            onFiltersChange({ ...filters, tags: event.target.value.split(',') })
        },
        [filters, onFiltersChange]
    )

    return (
        <Form>
            <div className="form-group">
                <label htmlFor="component-filters-form__query" className="sr-only">
                    Query
                </label>
                <input
                    id="component-filters-form__query"
                    className="form-control flex-grow-1"
                    type="search"
                    placeholder="Search..."
                    value={filters.query}
                    onChange={onQueryChange}
                />
            </div>
            <div className="form-group">
                <label htmlFor="component-filters-form__owner">Owner</label>
                <input
                    id="component-filters-form__owner"
                    className="form-control flex-grow-1"
                    value={filters.owner || ''}
                    onChange={onOwnerChange}
                />
            </div>
            <div className="form-group">
                <label htmlFor="component-filters-form__system">System</label>
                <input
                    id="component-filters-form__system"
                    className="form-control flex-grow-1"
                    value={filters.system || ''}
                    onChange={onSystemChange}
                />
            </div>
            <div className="form-group">
                <label htmlFor="component-filters-form__tags">Tags</label>
                <input
                    id="component-filters-form__tags"
                    className="form-control flex-grow-1"
                    value={filters.tags || ''}
                    onChange={onTagsChange}
                />
            </div>
        </Form>
    )
}

const FeedbackPopoverButton: React.FunctionComponent = () => {
    const buttonReference = useRef<HTMLButtonElement>(null)
    const [isVisible, setVisibility] = useState(false)

    return (
        <div className="d-flex align-items-center">
            <Badge status="wip" className="text-uppercase mr-2" />
            <Button ref={buttonReference} variant="link" size="sm">
                Share feedback
            </Button>
            <Popover
                isOpen={isVisible}
                target={buttonReference}
                onVisibilityChange={setVisibility}
                className={styles.feedbackPrompt}
            >
                <FeedbackPromptContent
                    closePrompt={() => setVisibility(false)}
                    textPrefix="Code Insights: "
                    routeMatch="/insights/dashboards"
                />
            </Popover>
        </div>
    )
}
