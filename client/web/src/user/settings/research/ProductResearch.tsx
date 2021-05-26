import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import React, { useEffect } from 'react'

import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../../../auth'
import { PageHeader } from '@sourcegraph/wildcard'

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
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Product research and feedback' }]}
                description={
                    <>
                        Our product team conducts occasional research to learn about how you use Sourcegraph and ask for
                        feedback about upcoming ideas. Sign up to participate in our research and help us shape the
                        future of our product!
                    </>
                }
                className="mb-3"
            />
            <a href={signUpForm.href} className="btn btn-primary mt-2" target="_blank" rel="noopener noreferrer">
                Sign up now <OpenInNewIcon className="icon-inline" />
            </a>
        </>
    )
}
