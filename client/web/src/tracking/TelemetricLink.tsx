import React, { useCallback, useState } from 'react'

import { LoaderButton } from '../components/LoaderButton'
import { logEventSynchronously } from '../user/settings/backend'

export interface TelemetricLinkProps {
    to: string
    label: string
    alwaysShowLabel: boolean
    eventName: string
    className?: string
}

const MAXIMUM_TELEMETRY_DELAY = 5000

export const TelemetricLink: React.FunctionComponent<React.PropsWithChildren<TelemetricLinkProps>> = ({
    to,
    label,
    alwaysShowLabel,
    eventName,
    className,
}) => {
    const [loading, setLoading] = useState(false)

    const onClick = useCallback(() => {
        if (loading) {
            return
        }

        setLoading(true)

        const navigate = (): void => {
            window.location.href = to
        }

        Promise.race([
            // Begin to log event
            logEventSynchronously(eventName),
            // If the event takes >5s, then we go ahead with the navigation unconditionally
            new Promise(resolve => setTimeout(resolve, MAXIMUM_TELEMETRY_DELAY)),
        ])
            .then(
                // Navigate unconditionally
                () => navigate(),
                () => navigate()
            )
            .then(
                () => setLoading(false),
                () => {}
            )
    }, [eventName, loading, to])

    return (
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
