import React from 'react'

import { buildGetStartedURL } from '@sourcegraph/shared/src/util/url'
import { ButtonLink } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'

export const CodeMonitorSignUpLink: React.FunctionComponent<
    React.PropsWithChildren<{
        className?: string
        text: string
        eventName: string
        cloudSignup?: boolean
    }>
> = ({ className, text, eventName, cloudSignup }) => {
    const onClick = (): void => {
        eventLogger.log(eventName)
    }
    return (
        <ButtonLink
            onClick={onClick}
            to={buildGetStartedURL(cloudSignup, '/code-monitoring/new')}
            className={className}
            variant="primary"
        >
            {text}
        </ButtonLink>
    )
}
