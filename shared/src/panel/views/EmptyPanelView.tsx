import CancelIcon from 'mdi-react/CancelIcon'
import React from 'react'

/**
 * An empty panel view.
 */
export const EmptyPanelView: React.FunctionComponent = () => (
    <div className="panel__empty">
        <CancelIcon className="icon-inline" /> Nothing to show here
    </div>
)
