import React, { useEffect } from 'react'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import { TelemetryService } from '../../../../../shared/src/telemetry/telemetryService'
import { AuthenticatedUser } from '../../../auth'

interface Props {
    telemetryService: TelemetryService
    authenticatedUser: Pick<AuthenticatedUser, 'email'>
}

const SIGN_UP_FORM_URL = 'https://info.sourcegraph.com/product-research'

export const ProductResearchPage: React.FunctionComponent<Props> = ({ telemetryService, authenticatedUser }) => {
    useEffect(() => {
        telemetryService.logViewEvent('UserSettingsProductResearch')
    }, [telemetryService])

    const signUpForm = new URL(SIGN_UP_FORM_URL)
    signUpForm.searchParams.set('email', authenticatedUser.email)

    return (
        <>
            <h2 className="mb-3">Product research and feedback</h2>
            <p>
                Our product team conducts occasional research to learn about how you use Sourcegraph and ask for
                feedback about upcoming ideas. Sign up to participate in our research and help us shape the future of
                our product!
            </p>
            <a href={signUpForm.href} className="btn btn-primary mt-2" target="_blank" rel="noopener noreferrer">
                Sign up now <OpenInNewIcon className="icon-inline" />
            </a>
        </>
    )
}
