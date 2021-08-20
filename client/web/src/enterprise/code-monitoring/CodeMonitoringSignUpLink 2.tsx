import classNames from 'classnames'
import React from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

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
        <Link
            onClick={onClick}
            to={`/sign-up?returnTo=${encodeURIComponent('/code-monitoring/new')}&src=Monitor`}
            className={classNames('btn btn-primary', className)}
        >
            {text}
        </Link>
    )
}
