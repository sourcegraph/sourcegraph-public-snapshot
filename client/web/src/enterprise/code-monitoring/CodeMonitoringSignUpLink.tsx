import React from 'react'

import { Button } from '@sourcegraph/wildcard'

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
            href={`https://about.sourcegraph.com/get-started?returnTo=${encodeURIComponent(
                '/code-monitoring/new'
            )}&src=Monitor`}
            className={className}
            variant="primary"
            as="a"
        >
            {text}
        </Button>
    )
}
