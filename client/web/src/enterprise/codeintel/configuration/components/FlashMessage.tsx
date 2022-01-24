import classNames from 'classnames'
import React from 'react'

interface FlashMessageProps {
    state: string
    message: string
    className?: string
}

export const FlashMessage: React.FunctionComponent<FlashMessageProps> = ({ state, message, className }) => (
    <div className={classNames('alert', className, state === 'SUCCESS' ? 'alert-success' : 'alert-warning')}>
        {message}
    </div>
)
