import React, { useEffect } from 'react'

import { mdiOpenInNew } from '@mdi/js'

import type { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader, ButtonLink, Icon, Text } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../../auth'
import { PageTitle } from '../../../components/PageTitle'

interface Props {
    telemetryService: TelemetryService
    authenticatedUser: Pick<AuthenticatedUser, 'emails'>
    isCodyApp: boolean
}

const SIGN_UP_FORM_URL = 'https://info.sourcegraph.com/product-research'

export const ProductResearchPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    telemetryService,
    authenticatedUser,
    isCodyApp,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('UserSettingsProductResearch')
    }, [telemetryService])

    const signUpForm = new URL(SIGN_UP_FORM_URL)
    const primaryEmail = authenticatedUser.emails.find(email => email.isPrimary)
    if (primaryEmail) {
        signUpForm.searchParams.set('email', primaryEmail.email)
    }

    return (
        <>
            <PageTitle title="Product research" />
            <PageHeader headingElement="h2" path={[{ text: 'Product research and feedback' }]} className="mb-3" />
            {isCodyApp && (
                <Container className="mb-2">
                    <Text>Do you have feedback or need help with Cody App?</Text>
                    {[
                        {
                            content: 'File an issue',
                            path: 'https://github.com/sourcegraph/app/issues/new?assignees=&labels=&template=bug_report.md&title=',
                            variant: 'primary' as const,
                        },
                        {
                            content: 'Join our Discord',
                            path: 'https://sourcegraph.com/community',
                        },
                    ].map(({ content, path, variant }) => (
                        <ButtonLink
                            key={path}
                            to={path}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="mr-2"
                            variant={variant ?? 'secondary'}
                        >
                            {content} <Icon aria-hidden={true} svgPath={mdiOpenInNew} />
                        </ButtonLink>
                    ))}
                </Container>
            )}
            <Container>
                <Text>
                    Our product team conducts occasional research to learn about how you use Sourcegraph and ask for
                    feedback about upcoming ideas. Sign up to participate in our research and help us shape the future
                    of our product!
                </Text>
                <ButtonLink to={signUpForm.href} target="_blank" rel="noopener noreferrer" variant="primary">
                    Sign up now <Icon aria-hidden={true} svgPath={mdiOpenInNew} />
                </ButtonLink>
            </Container>
        </>
    )
}
