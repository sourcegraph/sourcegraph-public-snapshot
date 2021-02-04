import React, { useEffect } from 'react'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import OpenInNew from 'mdi-react/OpenInNewIcon'

interface Props extends TelemetryProps {}

export const ProductResearchPage: React.FunctionComponent<Props> = props => {
    useEffect(() => {
        props.telemetryService.logViewEvent('UserSettingsProductResearch')
    }, [props.telemetryService])

    return (
        <>
            <h2 className="mb-3">Product research and feedback</h2>
            <div>
                Our product team conducts occasional research to learn about how you use Sourcegraph and ask for
                feedback about upcoming ideas. Sign up to participate in our research and help us shape the future of
                our product!
            </div>
            <a href="/somewhere" className="btn btn-primary mt-4">
                Sign up now <OpenInNew className="icon-inline" />
            </a>
        </>
    )
}
