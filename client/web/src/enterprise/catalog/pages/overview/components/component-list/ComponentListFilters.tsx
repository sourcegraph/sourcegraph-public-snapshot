import classNames from 'classnames'
import React, { ChangeEventHandler, useCallback, useRef } from 'react'
import { Form } from 'reactstrap'

import { CatalogComponentFiltersProps } from '../../../../core/component-filters'

import styles from './ComponentListFilters.module.scss'

interface Props extends CatalogComponentFiltersProps {
    size: 'sm' | 'lg'
    className?: string
}

export const ComponentListFilters: React.FunctionComponent<Props> = ({ filters, onFiltersChange, size, className }) => {
    // Update filter query on submit (not incrementally while typing).
    const queryElement = useRef<HTMLInputElement | null>(null)
    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        event => {
            event.preventDefault()
            onFiltersChange({ ...filters, query: queryElement.current?.value })
        },
        [filters, onFiltersChange]
    )

    // All other filters are updated upon change.
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

    const sizeStyles =
        size === 'sm'
            ? { form: styles.formSm, formGroup: styles.formGroupSm, input: undefined }
            : { form: styles.formLg, formGroup: styles.formGroupLg, input: undefined }

    return (
        <Form className={classNames(sizeStyles.form, className)} onSubmit={onSubmit}>
            <div className={classNames('form-group mb-0', sizeStyles.formGroup)}>
                <label htmlFor="component-list-filters__query" className="sr-only">
                    Query
                </label>
                <input
                    id="component-list-filters__query"
                    className={classNames('form-control', sizeStyles.input)}
                    type="search"
                    placeholder="Search..."
                    defaultValue={filters.query}
                    ref={queryElement}
                />
            </div>
            {size === 'lg' && (
                <>
                    <div className={classNames('form-group mb-0', sizeStyles.formGroup)}>
                        <label htmlFor="component-list-filters__owner" className="sr-only">
                            Owner
                        </label>
                        <input
                            id="component-list-filters__owner"
                            className={classNames('form-control', sizeStyles.input)}
                            placeholder="Owner"
                            value={filters.owner || ''}
                            onChange={onOwnerChange}
                        />
                    </div>
                    <div className={classNames('form-group mb-0', sizeStyles.formGroup)}>
                        <label htmlFor="component-list-filters__system" className="sr-only">
                            System
                        </label>
                        <input
                            id="component-list-filters__system"
                            className={classNames('form-control', sizeStyles.input)}
                            placeholder="System"
                            value={filters.system || ''}
                            onChange={onSystemChange}
                        />
                    </div>
                    <div className={classNames('form-group mb-0', sizeStyles.formGroup)}>
                        <label htmlFor="component-list-filters__tags" className="sr-only">
                            Tags
                        </label>
                        <input
                            id="component-list-filters__tags"
                            className={classNames('form-control', sizeStyles.input)}
                            placeholder="Tags"
                            value={filters.tags || ''}
                            onChange={onTagsChange}
                        />
                    </div>
                </>
            )}
            <button type="submit" className="sr-only">
                Filter
            </button>
        </Form>
    )
}
