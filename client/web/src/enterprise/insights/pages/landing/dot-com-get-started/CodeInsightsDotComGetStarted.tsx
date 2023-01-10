import React, { useEffect } from 'react'

import classNames from 'classnames'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { buildCloudTrialURL } from '@sourcegraph/shared/src/util/url'
import { Button, Card, CardBody, Link, PageHeader } from '@sourcegraph/wildcard'

import { CloudCtaBanner } from '../../../../../components/CloudCtaBanner'
import { Page } from '../../../../../components/Page'
import { PageTitle } from '../../../../../components/PageTitle'
import { CodeInsightsIcon } from '../../../../../insights/Icons'
import { eventLogger } from '../../../../../tracking/eventLogger'
import { CodeInsightsLandingPageContext, CodeInsightsLandingPageType } from '../CodeInsightsLandingPageContext'
import { CodeInsightsDescription } from '../getting-started/components/code-insights-description/CodeInsightsDescription'

import { CodeInsightsExamplesPicker } from './components/code-insights-examples-picker/CodeInsightsExamplesPicker'

import styles from './CodeInsightsDotComGetStarted.module.scss'

const DOT_COM_CONTEXT = { mode: CodeInsightsLandingPageType.Cloud }

export interface CodeInsightsDotComGetStartedProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
}

export const CodeInsightsDotComGetStarted: React.FunctionComponent<
    React.PropsWithChildren<CodeInsightsDotComGetStartedProps>
> = props => {
    const { telemetryService } = props

    useEffect(() => {
        telemetryService.logViewEvent('CloudInsightsGetStartedPage')
    }, [telemetryService])

    return (
        <CodeInsightsLandingPageContext.Provider value={DOT_COM_CONTEXT}>
            <Page>
                <PageTitle title="Code Insights" />
                <PageHeader
                    path={[{ icon: CodeInsightsIcon, text: 'Insights' }]}
                    actions={
                        <Button
                            as={Link}
                            to={buildCloudTrialURL(props.authenticatedUser, 'insights')}
                            target="_blank"
                            rel="noopener noreferrer"
                            variant="primary"
                            onClick={() => eventLogger.log('ClickedOnCloudCTA', { cloudCtaType: 'TryInsights' })}
                        >
                            Try insights
                        </Button>
                    }
                    className="mb-4"
                />
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

                    <CloudCtaBanner variant="filled">
                        To track Insights across your team's private repos,{' '}
                        <Link
                            to={buildCloudTrialURL(props.authenticatedUser, 'insights')}
                            target="_blank"
                            rel="noopener noreferrer"
                            onClick={() => eventLogger.log('ClickedOnCloudCTA', { cloudCtaType: 'Insights' })}
                        >
                            try Sourcegraph Cloud
                        </Link>
                        .
                    </CloudCtaBanner>

                    <CodeInsightsExamplesPicker telemetryService={telemetryService} />
                </main>
            </Page>
        </CodeInsightsLandingPageContext.Provider>
    )
}
