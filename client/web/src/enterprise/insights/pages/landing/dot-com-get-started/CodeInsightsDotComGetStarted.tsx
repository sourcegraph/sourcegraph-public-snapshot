import React, { useEffect } from 'react'

import { mdiArrowRight } from '@mdi/js'
import classNames from 'classnames'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Card, CardBody, Icon, Link, PageHeader } from '@sourcegraph/wildcard'

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

    return (
        <CodeInsightsLandingPageContext.Provider value={DOT_COM_CONTEXT}>
            <Page>
                <PageTitle title="Code Insights" />
                <PageHeader
                    path={[{ icon: CodeInsightsIcon, text: 'Insights' }]}
                    actions={
                        <Button
                            as={Link}
                            to="https://signup.sourcegraph.com/?p=insights"
                            variant="primary"
                            onClick={() => telemetryService.log('ClickedOnCloudCTA', { url: window.location.href })}
                        >
                            Try Insights
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
                    
                    <section className="my-3 p-2 d-flex justify-content-center bg-primary-4">
                        <Icon className="mr-2 text-merged" size="md" aria-hidden={true} svgPath={mdiArrowRight} />
                        <p className="mb-0">
                            To track Insights across your team's private repos,{' '}
                            <Link to="https://signup.sourcegraph.com/?p=insights" onClick={() => telemetryService.log('ClickedOnCloudCTA', { url: window.location.href })}>
                                try Sourcegraph Cloud
                            </Link>.
                        </p>
                    </section>

                    <CodeInsightsExamplesPicker telemetryService={telemetryService} />
                </main>
            </Page>
        </CodeInsightsLandingPageContext.Provider>
    )
}
