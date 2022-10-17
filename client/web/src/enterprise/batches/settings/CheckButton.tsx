import React from 'react'

import { mdiCheck, mdiClose } from '@mdi/js'

import { Button, LoadingSpinner, Icon } from '@sourcegraph/wildcard'

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
                <Icon svgPath={mdiCheck} inline={false} aria-label="Success" /> {successMessage}
            </div>
        )
    }
    if (failedMessage) {
        return (
            <div className="text-danger">
                <Icon svgPath={mdiClose} inline={false} aria-label="Failed" /> {failedMessage}
            </div>
        )
    }
    throw new Error('unreachable check button state')
}
