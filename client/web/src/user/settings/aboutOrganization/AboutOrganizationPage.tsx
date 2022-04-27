import React, { useEffect } from 'react'

import OpenInNewIcon from 'mdi-react/OpenInNewIcon'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader, ButtonLink, Icon } from '@sourcegraph/wildcard'

import { PageTitle } from '../../../components/PageTitle'
import { SelfHostedCta } from '../../../components/SelfHostedCta'

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
                <ButtonLink
                    to="https://share.hsforms.com/14OQ3RoPpQTOXvZlUpgx6-A1n7ku?utm_medium=direct-traffic&utm_source=in-product&utm_term=in-product-settings&utm_content=cloud-product-beta-teams"
                    target="_blank"
                    rel="noopener noreferrer"
                    variant="primary"
                >
                    Sign up for private beta access <Icon as={OpenInNewIcon} />
                </ButtonLink>
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
