import React from 'react'

import { buildGetStartedURL } from '@sourcegraph/shared/src/util/url'
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
            href={buildGetStartedURL('code-monitoring', '/code-monitoring/new')}
            className={className}
            variant="primary"
            as="a"
        >
            {text}
        </Button>
    )
}
