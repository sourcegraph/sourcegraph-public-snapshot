import React, { useEffect } from 'react'

import classNames from 'classnames'

import { DownloadSourcegraphIcon } from '@sourcegraph/branded/src/components/DownloadSourcegraphIcon'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Button,
    Card,
    CardBody,
    Link,
    PageHeader,
    H2,
    H3,
    Text,
    FeedbackPromptAuthenticatedUserProps,
} from '@sourcegraph/wildcard'

import { CtaBanner } from '../../../../../components/CtaBanner'
import { Page } from '../../../../../components/Page'
import { PageTitle } from '../../../../../components/PageTitle'
import { CodeInsightsIcon } from '../../../../../insights/Icons'
import { CodeInsightsLandingPageContext, CodeInsightsLandingPageType } from '../CodeInsightsLandingPageContext'
import { CodeInsightsLearnMore } from '../getting-started/components/code-insights-learn-more/CodeInsightsLearnMore'
import { CodeInsightsTemplates } from '../getting-started/components/code-insights-templates/CodeInsightsTemplates'

import { CodeInsightsExamplesPicker } from './components/code-insights-examples-picker/CodeInsightsExamplesPicker'

import styles from './CodeInsightsDotComGetStarted.module.scss'

const DOT_COM_CONTEXT = { mode: CodeInsightsLandingPageType.Cloud }

export interface CodeInsightsDotComGetStartedProps extends TelemetryProps, FeedbackPromptAuthenticatedUserProps {}

export const CodeInsightsDotComGetStarted: React.FunctionComponent<
    React.PropsWithChildren<CodeInsightsDotComGetStartedProps>
> = props => {
    const { telemetryService, authenticatedUser } = props

    useEffect(() => {
        telemetryService.logViewEvent('CloudInsightsGetStartedPage')
    }, [telemetryService])

    const handleScheduleDemoClick = (): void => {
        telemetryService.log('CloudCodeInsightsGetStartedScheduleDemo')
    }

    const handleExploreUseCasesClick = (): void => {
        telemetryService.log('CloudCodeInsightsGetStartedExploreUseCases')
    }

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

                        <section className={styles.hereDescriptionBlock}>
                            <H2 className={classNames(styles.heroTitle)}>
                                Draw insights from your codebase about how different initiatives are tracking over time
                            </H2>

                            <Text>
                                Create customizable, visual dashboards with meaningful codebase signals your team can
                                use to answer questions about what's in your code and how your code is changing.
                                Anything you can search, you can create a Code Insight for.
                            </Text>

                            <H3 className={classNames(styles.hereBulletTitle)}>Use Code Insights to...</H3>

                            <ul>
                                <li>Track migrations, adoption, and deprecations</li>
                                <li>Detect versions of languages, packages, or infrastructure</li>
                                <li>Ensure removal of security vulnerabilities</li>
                                <li>Track code smells, ownership, and configurations</li>
                            </ul>

                            <footer className={styles.heroFooter}>
                                <Button
                                    variant="primary"
                                    as={Link}
                                    to="https://about.sourcegraph.com/contact/request-code-insights-demo?utm_medium=direct-traffic&utm_source=in-product&utm_campaign=code-insights-getting-started"
                                    target="_blank"
                                    rel="noopener"
                                    onClick={handleScheduleDemoClick}
                                >
                                    Schedule a demo
                                </Button>

                                <Button
                                    variant="secondary"
                                    as={Link}
                                    to="/help/code_insights/references/common_use_cases"
                                    onClick={handleExploreUseCasesClick}
                                >
                                    Explore use cases
                                </Button>
                            </footer>
                        </section>
                    </Card>

                    <section className={styles.quoteSection}>
                        <H2>Trusted by leading engineering teams around the world:</H2>

                        <q className={styles.quote}>
                            As we’ve grown, so has the need to better track and communicate our progress and goals
                            across the engineering team and broader company. With Code Insights, our data and migration
                            tracking is accurate across our entire codebase, and our engineers and managers can shift
                            out of manual spreadsheets and spend more time working on code.
                        </q>

                        <span className={styles.quoteAuthor}>Balázs Tóthfalussy, Engineering Manager</span>

                        <img
                            className={styles.quoteLogo}
                            width={82}
                            height={30}
                            src="https://storage.googleapis.com/sourcegraph-assets/code_insights/prezi-logo-lg.png"
                            alt="Prezi logotype"
                        />
                    </section>

                    <CodeInsightsExamplesPicker telemetryService={telemetryService} />

                    <section className={styles.ctaSection}>
                        <CtaBanner
                            bodyText="Code Insights requires a Sourcegraph Cloud or self-hosted instance."
                            title={<H3>Start using Code Insights</H3>}
                            linkText="Get started"
                            href="/help/admin/install?utm_medium=direct-traffic&utm_source=in-product&utm_campaign=code-insights-getting-started"
                            icon={<DownloadSourcegraphIcon />}
                            onClick={handleInstallLocalInstanceClick}
                        />
                    </section>

                    <CodeInsightsTemplates className={styles.templateSection} telemetryService={telemetryService} />

                    <iframe
                        src="https://www.youtube.com/embed/fMCUJQHfbUA"
                        title="YouTube video player"
                        frameBorder="0"
                        allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                        allowFullScreen={true}
                        className={styles.videoSection}
                    />

                    <CodeInsightsLearnMore
                        className={styles.learnMoreSection}
                        telemetryService={telemetryService}
                        authenticatedUser={authenticatedUser}
                    />
                </main>
            </Page>
        </CodeInsightsLandingPageContext.Provider>
    )
}
