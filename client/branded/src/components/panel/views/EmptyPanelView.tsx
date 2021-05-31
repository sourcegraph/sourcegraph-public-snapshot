import classNames from 'classnames'
import CancelIcon from 'mdi-react/CancelIcon'
import React from 'react'

interface EmptyPanelViewProps {
    className?: string
}

export const EmptyPanelView: React.FunctionComponent<EmptyPanelViewProps> = props => {
    const { className } = props

    return (
        <div className={classNames('panel__empty', className)}>
            <CancelIcon className="icon-inline" /> Nothing to show here
        </div>
    )
}
