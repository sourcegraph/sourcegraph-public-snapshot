import React, { useCallback, useState } from 'react'

import { Redirect } from 'react-router'

import { LoaderButton } from '../components/LoaderButton'
import { logEventSynchronously } from '../user/settings/backend'

export interface TelemetricRedirectProps {
    to: string
    label: string
    alwaysShowLabel: boolean
    eventName: string
    className?: string
}

const MAXIMUM_TELEMETRY_DELAY = 5000

export const TelemetricRedirect: React.FunctionComponent<React.PropsWithChildren<TelemetricRedirectProps>> = ({
    to,
    label,
    alwaysShowLabel,
    eventName,
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
            new Promise(resolve => setTimeout(resolve, MAXIMUM_TELEMETRY_DELAY)),
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
        <LoaderButton
            variant="link"
            label={label}
            alwaysShowLabel={alwaysShowLabel}
            className={className}
            onClick={onClick}
            loading={loading}
        />
    )
}
