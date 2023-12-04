import React from 'react'

import { mdiCheck, mdiClose } from '@mdi/js'

import { Button, LoadingSpinner, Icon, Alert } from '@sourcegraph/wildcard'

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
            <Button
                className="text-primary text-nowrap font-weight-normal"
                onClick={onClick}
                variant="link"
                size="sm"
                aria-label={label}
            >
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
            <Alert className="text-success m-0 small">
                <Icon svgPath={mdiCheck} inline={false} aria-hidden={true} /> {successMessage}
            </Alert>
        )
    }
    if (failedMessage) {
        return (
            <Alert className="text-danger m-0 small">
                <Icon svgPath={mdiClose} inline={false} aria-hidden={true} /> {failedMessage}
            </Alert>
        )
    }
    throw new Error('unreachable check button state')
}
