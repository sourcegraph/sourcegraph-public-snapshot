import { useEffect } from 'react'

import { mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'

import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { addSourcegraphAppOutboundUrlParameters } from '@sourcegraph/shared/src/util/url'
import { Container, H1, H2, Icon, Link, PageHeader, ProductStatusBadge, Text } from '@sourcegraph/wildcard'

import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'

import styles from './AppComingSoonPage.module.scss'

export const AppComingSoonPage: React.FC = () => {
    const isLightTheme = useIsLightTheme()
    useEffect(() => eventLogger.logPageView('AppComingSoonPage'), [])

    return (
        <Page className={styles.root}>
            <PageTitle title="Coming soon" />
            <PageHeader
                description="Exciting things are coming to Sourcegraph. Stay tuned for upcoming features."
                className="mb-3"
            >
                <H1 as="h2" className="d-flex align-items-center">
                    Coming soon
                </H1>
            </PageHeader>

            <Container className={classNames('mb-3', styles.container)}>
                <section className={classNames('row', styles.section)}>
                    <div className={classNames('col-12 col-md-5', styles.text)}>
                        <ProductStatusBadge status="experimental" className="mb-4 text-uppercase" />
                        <H2 as="h3" className="mb-4">
                            Cody
                        </H2>
                        <Text className="mb-4">
                            <strong>Read, write, and understand code 10x faster with AI</strong>
                            <br />
                            Your intelligent, code-aware, enterprise-ready programmer’s assistant.
                        </Text>
                        <Text className="mb-4">
                            <strong>Codebase-aware chat</strong>
                            <br />
                            Answer questions about both general programming topics and your specific codebase from right
                            inside your editor. Cody knows about your local code and can learn from all the code and
                            documentation inside your organization.
                        </Text>
                        <Link
                            className={styles.link}
                            to={addSourcegraphAppOutboundUrlParameters('https://about.sourcegraph.com/cody', 'cody')}
                            target="_blank"
                            rel="noopener"
                        >
                            Learn more <Icon role="img" aria-label="Open in a new tab" svgPath={mdiChevronRight} />
                        </Link>
                    </div>
                    <div className={classNames('col-12 col-md-7', styles.imageWrapper)}>
                        <img
                            src="https://storage.googleapis.com/sourcegraph-assets/app-coming-soon-cody.png"
                            alt="Screenshot of the Cody VS Code extension. You can see it being asked the question 'Can you show me another example of using an interface from my codebase?' and it responds with a concrete TypeScript interface example from the codebase."
                            width={612}
                            className={classNames('max-w-100 max-h-100 percy-hide', styles.image)}
                        />
                    </div>
                </section>
            </Container>

            <Container className={classNames('mb-3', styles.container)}>
                <section className={classNames('row', styles.section)}>
                    <div className={classNames('col-12 col-md-5', styles.text)}>
                        <ProductStatusBadge status="experimental" className="mb-4 text-uppercase" />
                        <H2 as="h3" className="mb-4">
                            Sourcegraph Own
                        </H2>
                        <Text className="mb-4">
                            <strong>100% code ownership coverage, now</strong> <br />
                            Evergreen code ownership across code hosts, powering Code Search, Batch Changes, and
                            Insights.
                        </Text>
                        <Text className="mb-4">
                            <strong>Resolve incidents and security vulnerabilities faster</strong>
                            <br />
                            Search for vulnerable or outdated code patterns and reach out to the owners in seconds.
                            Escalate in one click. Don’t waste time emailing around to find who can help. Fast
                            collaboration, fast remediation.
                        </Text>
                        <Link
                            className={styles.link}
                            to={addSourcegraphAppOutboundUrlParameters('https://about.sourcegraph.com/own', 'own')}
                            target="_blank"
                            rel="noopener"
                        >
                            Learn more <Icon role="img" aria-label="Open in a new tab" svgPath={mdiChevronRight} />
                        </Link>
                    </div>
                    <div className={classNames('col-12 col-md-7', styles.imageWrapper)}>
                        <img
                            src={`https://storage.googleapis.com/sourcegraph-assets/app-coming-soon-own-${
                                isLightTheme ? 'light' : 'dark'
                            }.png`}
                            alt="Screenshot of the Own features on the Sourcegraph web app. It shows a new tab in the bottom panel on the code view that lists ownership information for a specific file and classifies owners into exports, how often they made changes, and how they were categorized."
                            width={612}
                            className={classNames('max-w-100 percy-hide', styles.image)}
                        />
                    </div>
                </section>
            </Container>
        </Page>
    )
}
