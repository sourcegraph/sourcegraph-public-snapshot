import React, { useEffect } from 'react'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'

interface Props extends TelemetryProps {}

export const ProductFeedbackPage: React.FunctionComponent<Props> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.logViewEvent('UserSettingsRepositories')
    }, [telemetryService])

    return (
        <>
            <h2>Product research and feedback</h2>
            <div>
                Our product team conducts occasional research to learn about how you use Sourcegraph and ask for
                feedback about upcoming ideas. Sign up to participate in our research and help us shape the future of
                our product!Î
            </div>
            {/* TODO: Add button icon */}
            <a href="/somewhere" className="btn btn-primary">
                Sign up now
            </a>
        </>
    )
}
