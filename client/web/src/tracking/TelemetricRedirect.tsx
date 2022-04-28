import React, { useCallback, useState } from 'react'

import { Redirect } from 'react-router'

import { LoaderButton } from '../components/LoaderButton'
import { logEventSynchronously } from '../user/settings/backend'

export interface TelemetricRedirectProps {
    to: string
    eventName: string
    children: React.ReactNode
    className?: string
}

const MaximumTelemetryDelay = 5000

export const TelemetricRedirect: React.FunctionComponent<TelemetricRedirectProps> = ({
    to,
    eventName,
    children,
    className,
}) => {
    const [loading, setLoading] = useState(false)
    const [redirect, setRedirect] = useState(false)

    const onClick = useCallback(() => {
        if (loading) {
            return
        }

        setLoading(true)

        Promise.race([
            // Begin to log event
            logEventSynchronously(eventName),
            // If the event takes >5s, then we go ahead with the redirect
            new Promise(resolve => setTimeout(resolve, MaximumTelemetryDelay)),
        ])
            .then(
                // Redirect unconditionally
                () => setRedirect(true),
                () => setRedirect(true)
            )
            .then(
                () => setLoading(false),
                () => {}
            )
    }, [setRedirect, eventName, loading])

    return redirect ? (
        <Redirect to={to} />
    ) : (
        <LoaderButton variant="link" className={className} onClick={onClick} loading={loading}>
            {children}
        </LoaderButton>
    )
}
