import React from 'react'
import { ConnectionListContext } from './ConnectionList'

export interface ConnectionListHeaderItems {
    /**
     * A React fragment to render on the left side of the list header.
     */
    left?: React.ReactFragment

    /**
     * A React fragment to render on the right side of the list header.
     */
    right?: React.ReactFragment
}

interface Props extends ConnectionListContext, ConnectionListContext {
    items?: ConnectionListHeaderItems
}

/**
 * The header for the {@link ConnectionList} component.
 */
export const ConnectionListHeader: React.FunctionComponent<Props> = ({ items, itemCheckboxes }) => (
    <div className="card-header d-flex align-items-center justify-content-between font-weight-normal">
        {itemCheckboxes && (
            <input
                type="checkbox"
                className="form-check mx-1 my-2"
                style={{ verticalAlign: 'middle' }}
                aria-label="Select item"
            />
        )}
        {items && items.left}
        <div className="flex-1" />
        {items && items.right}
    </div>
)
