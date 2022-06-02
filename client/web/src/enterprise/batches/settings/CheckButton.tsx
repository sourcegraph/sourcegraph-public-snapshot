import React from 'react'

import CheckIcon from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'

import { Button, LoadingSpinner } from '@sourcegraph/wildcard'

export interface CheckButtonProps {
    label: string
    onClick: React.MouseEventHandler
    loading: boolean
    successMessage?: string
    failedMessage?: string
}

export const CheckButton: React.FunctionComponent<React.PropsWithChildren<CheckButtonProps>> = ({
    label,
    onClick,
    loading,
    successMessage,
    failedMessage,
}) => {
    if (!loading && !successMessage && !failedMessage) {
        return (
            <Button className="text-primary text-nowrap" onClick={onClick} variant="link" aria-label={label}>
                Check
            </Button>
        )
    }
    if (loading) {
        return (
            <div className="text-muted">
                <LoadingSpinner /> Checking
            </div>
        )
    }
    if (successMessage && !failedMessage) {
        return (
            <div className="text-success">
                <CheckIcon /> {successMessage}
            </div>
        )
    }
    if (failedMessage) {
        return (
            <div className="text-danger">
                <CloseIcon /> {failedMessage}
            </div>
        )
    }
    throw new Error('unreachable check button state')
}
