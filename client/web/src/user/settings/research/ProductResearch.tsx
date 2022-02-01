import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import React, { useEffect } from 'react'

import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader, Button } from '@sourcegraph/wildcard'

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
            <PageHeader headingElement="h2" path={[{ text: 'Product research and feedback' }]} className="mb-3" />
            <Container>
                <p>
                    Our product team conducts occasional research to learn about how you use Sourcegraph and ask for
                    feedback about upcoming ideas. Sign up to participate in our research and help us shape the future
                    of our product!
                </p>
                <Button href={signUpForm.href} target="_blank" rel="noopener noreferrer" variant="primary" as="a">
                    Sign up now <OpenInNewIcon className="icon-inline" />
                </Button>
            </Container>
        </>
    )
}
