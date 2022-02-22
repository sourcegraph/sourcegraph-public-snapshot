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
                <section className={styles.chartSection}>
                    <div className={styles.heroImage} />
                </section>

                <section>
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
                        <Button variant="primary" as={Link} to="/schedule-demo">
                            Schedule a demo
                        </Button>
                        <Button variant="secondary" as={Link} to="/schedule-demo">
                            Explore use cases
                        </Button>
                    </footer>
                </section>
            </Card>

            <section className={styles.quoteSection}>
                <h2>Trusted by leading engineering teams around the world:</h2>

                <q className={styles.quote}>
                    Code insights enables our team to move away from manual spreadsheets and point-in-time documentation
                    and provides us with a holistic view of our codebase when we undergo complex projects such as
                    migrations and major platform-related changes.
                </q>

                <span className={styles.quoteAuthor}>Jane Doe, Engineering leader</span>

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

                        <Button as={Link} variant="primary" to="/install">
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
