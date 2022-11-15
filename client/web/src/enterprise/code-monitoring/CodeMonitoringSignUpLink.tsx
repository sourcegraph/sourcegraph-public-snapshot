import React from 'react'

import { buildGetStartedURL } from '@sourcegraph/shared/src/util/url'
import { ButtonLink } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'

export const CodeMonitorSignUpLink: React.FunctionComponent<
    React.PropsWithChildren<{
        className?: string
        text: string
        eventName: string
        forDotcom?: boolean
    }>
> = ({ className, text, eventName, forDotcom }) => {
    const onClick = (): void => {
        eventLogger.log(eventName)
    }
    return (
        <ButtonLink
            onClick={onClick}
            to={buildGetStartedURL(forDotcom, '/code-monitoring/new')}
            className={className}
            variant="primary"
        >
            {text}
        </ButtonLink>
    )
}
