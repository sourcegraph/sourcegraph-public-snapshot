import classNames from 'classnames'
import React from 'react'

interface FlashMessageProps {
    state: string
    message: string
}

export const FlashMessage: React.FunctionComponent<FlashMessageProps> = ({ state, message }) => (
    <div className={classNames('alert', state === 'SUCCESS' ? 'alert-success' : 'alert-warning')}>{message}</div>
)
