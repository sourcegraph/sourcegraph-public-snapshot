import React, { useRef } from 'react'

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
    const buttonReference = useRef<HTMLButtonElement | null>(null)
    return (
        <>
            {!loading && !successMessage && !failedMessage && (
                <Button
                    className="text-primary text-nowrap"
                    onClick={onClick}
                    variant="link"
                    aria-label={label}
                    ref={buttonReference}
                >
                    Check
                </Button>
            )}
            {loading && (
                <div className="text-white-50">
                    <LoadingSpinner /> Checking
                </div>
            )}
            {successMessage && !failedMessage && (
                <div className="text-success">
                    <CheckIcon /> {successMessage}
                </div>
            )}
            {failedMessage && (
                <div className="text-danger">
                    <CloseIcon /> {failedMessage}
                </div>
            )}
        </>
    )
}
