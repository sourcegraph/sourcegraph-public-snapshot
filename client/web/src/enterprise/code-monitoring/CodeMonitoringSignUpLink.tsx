import React from 'react'

import { Link, Button } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'

export const CodeMonitorSignUpLink: React.FunctionComponent<{
    className?: string
    text: string
    eventName: string
}> = ({ className, text, eventName }) => {
    const onClick = (): void => {
        eventLogger.log(eventName)
    }
    return (
        <Button
            onClick={onClick}
            to={`/sign-up?returnTo=${encodeURIComponent('/code-monitoring/new')}&src=Monitor`}
            className={className}
            variant="primary"
            as={Link}
        >
            {text}
        </Button>
    )
}
