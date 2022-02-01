import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { SelfHostedCta } from '@sourcegraph/web/src/components/SelfHostedCta'
import { Container, PageHeader, Button } from '@sourcegraph/wildcard'

import styles from './AboutOrganizationPage.module.scss'
interface AboutOrganizationPageProps extends TelemetryProps {}

export const AboutOrganizationPage: React.FunctionComponent<AboutOrganizationPageProps> = ({ telemetryService }) => {
    useEffect(() => {
        telemetryService.logViewEvent('AboutOrg')
    }, [telemetryService])

    return (
        <>
            <PageTitle title="Organizations" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Organizations' }]}
                description="Support for organizations is not currently available on Sourcegraph Cloud."
                className="mb-3"
            />
            <Container className="mb-4">
                <h3>Private beta access for small teams now available</h3>
                <p>
                    Get instant access to code navigation and intelligence across your team’s private code and 2M open
                    source repositories. Sourcegraph Cloud for teams brings enterprise advantages to small teams.
                </p>
                <Button
                    href="https://share.hsforms.com/14OQ3RoPpQTOXvZlUpgx6-A1n7ku/"
                    target="_blank"
                    rel="noopener noreferrer"
                    variant="primary"
                    as="a"
                >
                    Sign up for private beta access <OpenInNewIcon className="icon-inline" />
                </Button>
            </Container>
            <SelfHostedCta
                contentClassName={styles.selfHostedCtaContent}
                page="organizations"
                telemetryService={telemetryService}
            >
                <p className="mb-2">
                    <strong>Need more enterprise features? Run Sourcegraph self-hosted</strong>
                </p>
                <p className="mb-2">
                    For additional code hosts and enterprise only features, install Sourcegraph self-hosted.
                </p>
            </SelfHostedCta>
        </>
    )
}
