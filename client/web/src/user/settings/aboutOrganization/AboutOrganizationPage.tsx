import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import { SelfHostedCta } from '@sourcegraph/web/src/components/SelfHostedCta'
import { Container, PageHeader } from '@sourcegraph/wildcard'

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
            <Container>
                <SelfHostedCta
                    contentClassName={styles.selfHostedCtaContent}
                    page="organizations"
                    telemetryService={telemetryService}
                >
                    <p className="mb-2">
                        <strong>Run Sourcegraph self-hosted for more enterprise features</strong>
                    </p>
                    <p className="mb-2">
                        For team oriented functionality, additional code hosts and enterprise only features, install
                        Sourcegraph self-hosted.
                    </p>
                </SelfHostedCta>
            </Container>
        </>
    )
}
