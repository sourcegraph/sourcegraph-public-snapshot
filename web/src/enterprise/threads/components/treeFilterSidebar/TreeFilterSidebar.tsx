import React, { useCallback } from 'react'
import { Form } from '../../../../components/Form'
import { QueryParameterProps } from '../withQueryParameter/WithQueryParameter'

interface RenderChildrenProps extends Pick<QueryParameterProps, 'query'> {
    className?: string
}

interface Props extends QueryParameterProps {
    children: (props: RenderChildrenProps) => JSX.Element | null

    className?: string
}

/**
 * A sidebar containing a tree with a filter input field.
 */
export const TreeFilterSidebar: React.FunctionComponent<Props> = ({
    query,
    onQueryChange,
    children,
    className = '',
}) => {
    const onChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            onQueryChange(e.currentTarget.value)
        },
        [onQueryChange]
    )
    const onSubmit = useCallback<React.FormEventHandler>(e => {
        e.preventDefault()
    }, [])

    const ITEM_CLASS_NAME = 'list-group-item list-group-item-action small'
    return (
        <aside className={`overflow-hidden d-flex flex-column ${className}`}>
            <Form className="form p-2 flex-0" onSubmit={onSubmit}>
                <input
                    type="search"
                    className="form-control form-control-sm"
                    placeholder="Filter..."
                    value={query}
                    onChange={onChange}
                />
            </Form>
            <div className="list-group list-group-flush border-bottom" style={{ overflowY: 'auto' }}>
                {children({ query, className: ITEM_CLASS_NAME })}
            </div>
        </aside>
    )
}
