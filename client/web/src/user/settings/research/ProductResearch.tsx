import React, { useEffect } from 'react'
import OpenInNew from 'mdi-react/OpenInNewIcon'
import { TelemetryService } from '../../../../../shared/src/telemetry/telemetryService'
import { AuthenticatedUser } from '../../../auth'

interface Props {
    telemetryService: TelemetryService
    authenticatedUser: AuthenticatedUser
}

const PRODUCT_RESEARCH_SIGNUP_FORM = 'https://share.hsforms.com/1tkScUc65Tm-Yu98zUZcLGw1n7ku'

export const ProductResearchPage: React.FunctionComponent<Props> = ({ telemetryService, authenticatedUser }) => {
    useEffect(() => {
        telemetryService.logViewEvent('UserSettingsProductResearch')
    }, [telemetryService])

    return (
        <>
            <h2 className="mb-3">Product research and feedback</h2>
            <div>
                Our product team conducts occasional research to learn about how you use Sourcegraph and ask for
                feedback about upcoming ideas. Sign up to participate in our research and help us shape the future of
                our product!
            </div>
            <a
                href={`${PRODUCT_RESEARCH_SIGNUP_FORM}?email=${authenticatedUser.email}`}
                className="btn btn-primary mt-4"
                target="_blank"
                rel="noopener noreferrer"
            >
                Sign up now <OpenInNew className="icon-inline" />
            </a>
        </>
    )
}
