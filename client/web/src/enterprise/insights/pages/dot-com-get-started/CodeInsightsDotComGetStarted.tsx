import classNames from 'classnames'
import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Card, CardBody, Link, PageHeader } from '@sourcegraph/wildcard'

import { Page } from '../../../../components/Page'
import { PageTitle } from '../../../../components/PageTitle'
import { CodeInsightsIcon } from '../../../../insights/Icons'
import { CodeInsightsLearnMore } from '../getting-started/components/code-insights-learn-more/CodeInsightsLearnMore'
import { CodeInsightsTemplates } from '../getting-started/components/code-insights-templates/CodeInsightsTemplates'

import styles from './CodeInsightsDotComGetStarted.module.scss'
import { CodeInsightsExamplesPicker } from './components/code-insights-examples-picker/CodeInsightsExamplesPicker'
import { SourcegraphInstallLocallyIcon } from './components/SourcegraphInstallLocallyIcon'

export interface CodeInsightsDotComGetStartedProps extends TelemetryProps {}

export const CodeInsightsDotComGetStarted: React.FunctionComponent<CodeInsightsDotComGetStartedProps> = ({
    telemetryService,
}) => (
    <Page>
        <PageTitle title="Code Insights" />
        <PageHeader
            path={[{ icon: CodeInsightsIcon, text: 'Insights' }]}
            description="Code Insights description copy"
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

                <section className={styles.hereDescriptionBlock}>
                    <h2 className={classNames(styles.heroTitle)}>
                        Draw insights from your codebase about how different initiatives are tracking over time
                    </h2>

                    <p>
                        Create customizable, visual dashboards with meaningful codebase signals your team can use to
                        answer questions about what's in your code and how your code is changing. Anything you can
                        search, you can create a Code Insight for.
                    </p>

                    <h3 className={classNames(styles.hereBulletTitle)}>Use Code Insights to...</h3>

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
                            to="https://about.sourcegraph.com/contacts"
                            target="_blank"
                            rel="noopener"
                        >
                            Schedule a demo
                        </Button>

                        <Button variant="secondary" as={Link} to="/help/code_insights/references/common_use_cases">
                            Explore use cases
                        </Button>
                    </footer>
                </section>
            </Card>

            <section className={styles.quoteSection}>
                <h2>Trusted by leading engineering teams around the world:</h2>

                <q className={styles.quote}>
                    As we’ve grown, so has the need to better track and communicate our progress and goals across the
                    engineering team and broader company. With Code Insights, our data and migration tracking is
                    accurate across our entire codebase, and our engineers and managers can shift out of manual
                    spreadsheets and spend more time working on code.
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

            <section className={styles.installLocallySection}>
                <Card as={CardBody} className={styles.installLocallyRequirements}>
                    <h3>Code Insights requirements</h3>
                    <ul className={styles.installLocallyRequirementsList}>
                        <li>On-prem installation</li>
                        <li>Create up to 2 code insights for free</li>
                        <li>Get unlimited insights with an Enterprise plan (or trial)</li>
                    </ul>
                </Card>

                <Card as={CardBody} className={styles.installLocallyGetStarted}>
                    <SourcegraphInstallLocallyIcon className="flex-shrink-0" />
                    <div>
                        <h3>Install locally to get started</h3>

                        <p>
                            Code Insights requires a local Sourcegraph installation via Docker Compose or Kubernetes.
                            You can check it out for free by installing with a single line of code.
                        </p>

                        <Button
                            as={Link}
                            variant="primary"
                            to="/help/admin/install?utm_medium=inproduct&utm_source=inproduct-code-insights&term="
                        >
                            Install local instance
                        </Button>
                    </div>
                </Card>
            </section>

            <CodeInsightsTemplates
                interactive={false}
                className={styles.templateSection}
                telemetryService={telemetryService}
            />

            <section className={styles.videoSection}>Demo video placeholder</section>

            <CodeInsightsLearnMore className={styles.learnMoreSection} telemetryService={telemetryService} />
        </main>
    </Page>
)
