import classNames from 'classnames'
import React, { ChangeEventHandler, useCallback, useRef } from 'react'
import { Form } from 'reactstrap'

import { CatalogEntityFiltersProps } from '../../../../core/entity-filters'

import styles from './EntityListFilters.module.scss'

interface Props extends CatalogEntityFiltersProps {
    size: 'sm' | 'lg'
    className?: string
}

export const EntityListFilters: React.FunctionComponent<Props> = ({ filters, onFiltersChange, size, className }) => {
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
                <label htmlFor="entity-list-filters__query" className="sr-only">
                    Query
                </label>
                <input
                    id="entity-list-filters__query"
                    className={classNames('form-control', sizeStyles.input)}
                    type="search"
                    placeholder="Search catalog..."
                    defaultValue={filters.query}
                    ref={queryElement}
                />
            </div>
            {size === 'lg' && (
                <>
                    <div className={classNames('form-group mb-0', sizeStyles.formGroup)}>
                        <label htmlFor="entity-list-filters__owner" className="sr-only">
                            Owner
                        </label>
                        <input
                            id="entity-list-filters__owner"
                            className={classNames('form-control', sizeStyles.input)}
                            placeholder="Owner"
                            value={filters.owner || ''}
                            onChange={onOwnerChange}
                        />
                    </div>
                    <div className={classNames('form-group mb-0', sizeStyles.formGroup)}>
                        <label htmlFor="entity-list-filters__system" className="sr-only">
                            System
                        </label>
                        <input
                            id="entity-list-filters__system"
                            className={classNames('form-control', sizeStyles.input)}
                            placeholder="System"
                            value={filters.system || ''}
                            onChange={onSystemChange}
                        />
                    </div>
                    <div className={classNames('form-group mb-0', sizeStyles.formGroup)}>
                        <label htmlFor="entity-list-filters__tags" className="sr-only">
                            Tags
                        </label>
                        <input
                            id="entity-list-filters__tags"
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
