import React from 'react'


import { Button, LoadingSpinner, Icon } from '@sourcegraph/wildcard'
import { mdiCheck, mdiClose } from "@mdi/js";

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
                <Icon svgPath={mdiCheck} inline={false} aria-hidden={true} /> {successMessage}
            </div>
        )
    }
    if (failedMessage) {
        return (
            <div className="text-danger">
                <Icon svgPath={mdiClose} inline={false} aria-hidden={true} /> {failedMessage}
            </div>
        )
    }
    throw new Error('unreachable check button state')
}
