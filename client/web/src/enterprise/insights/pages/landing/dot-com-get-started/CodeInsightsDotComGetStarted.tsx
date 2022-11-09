import React, { useEffect } from 'react'

import classNames from 'classnames'

import { DownloadSourcegraphIcon } from '@sourcegraph/branded/src/components/DownloadSourcegraphIcon'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Card, CardBody, PageHeader, H3 } from '@sourcegraph/wildcard'

import { CtaBanner } from '../../../../../components/CtaBanner'
import { Page } from '../../../../../components/Page'
import { PageTitle } from '../../../../../components/PageTitle'
import { CodeInsightsIcon } from '../../../../../insights/Icons'
import { CodeInsightsLandingPageContext, CodeInsightsLandingPageType } from '../CodeInsightsLandingPageContext'
import { CodeInsightsDescription } from '../getting-started/components/code-insights-description/CodeInsightsDescription'

import { CodeInsightsExamplesPicker } from './components/code-insights-examples-picker/CodeInsightsExamplesPicker'

import styles from './CodeInsightsDotComGetStarted.module.scss'

const DOT_COM_CONTEXT = { mode: CodeInsightsLandingPageType.Cloud }

export interface CodeInsightsDotComGetStartedProps extends TelemetryProps {}

export const CodeInsightsDotComGetStarted: React.FunctionComponent<
    React.PropsWithChildren<CodeInsightsDotComGetStartedProps>
> = props => {
    const { telemetryService } = props

    useEffect(() => {
        telemetryService.logViewEvent('CloudInsightsGetStartedPage')
    }, [telemetryService])

    const handleInstallLocalInstanceClick = (): void => {
        telemetryService.log('CloudCodeInsightsGetStartedInstallInstance')
    }

    return (
        <CodeInsightsLandingPageContext.Provider value={DOT_COM_CONTEXT}>
            <Page>
                <PageTitle title="Code Insights" />
                <PageHeader path={[{ icon: CodeInsightsIcon, text: 'Insights' }]} className="mb-4" />
                <main className="pb-5">
                    <Card as={CardBody} className={styles.heroSection}>
                        <aside className={styles.heroVideoBlock}>
                            <video
                                className={classNames('shadow percy-hide w-100 h-auto')}
                                width={1280}
                                height={720}
                                autoPlay={true}
                                muted={true}
                                loop={true}
                                playsInline={true}
                                controls={false}
                            >
                                <source
                                    type="video/webm"
                                    src="https://storage.googleapis.com/sourcegraph-assets/code_insights/code-insights-720.webm"
                                />

                                <source
                                    type="video/mp4"
                                    src="https://storage.googleapis.com/sourcegraph-assets/code_insights/code-insights-720.mp4"
                                />
                            </video>
                        </aside>
                        <CodeInsightsDescription className={styles.heroDescriptionBlock} />
                    </Card>

                    <CodeInsightsExamplesPicker telemetryService={telemetryService} />

                    <section className="d-flex justify-content-start mt-3">
                        <CtaBanner
                            bodyText="Code Insights requires a Sourcegraph Cloud or self-hosted instance."
                            title={<H3>Start using Code Insights</H3>}
                            linkText="Get started"
                            href="/help/admin/install?utm_medium=direct-traffic&utm_source=in-product&utm_campaign=code-insights-getting-started"
                            icon={<DownloadSourcegraphIcon />}
                            onClick={handleInstallLocalInstanceClick}
                        />
                    </section>
                </main>
            </Page>
        </CodeInsightsLandingPageContext.Provider>
    )
}
