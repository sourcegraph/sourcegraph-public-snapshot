import { useEffect } from 'react'

import { mdiChevronRight } from '@mdi/js'
import classNames from 'classnames'

import {
    Container,
    H1,
    H2,
    Icon,
    Link,
    PageHeader,
    ProductStatusBadge,
    Text,
    useMatchMedia,
} from '@sourcegraph/wildcard'

import { Page } from '../../components/Page'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'

import styles from './AppComingSoonPage.module.scss'

export const AppComingSoonPage: React.FC = () => {
    useEffect(() => eventLogger.logPageView('CodeMonitoringGettingStartedPage'), [])
    const allowAutoplay = useMatchMedia('(prefers-reduced-motion: no-preference)')

    return (
        <Page>
            <PageTitle title="Coming soon" />
            <PageHeader
                description="Exciting things are coming to Sourcegraph. Stay tuned for upcoming features."
                className="mb-3"
            >
                <H1 as="h2" className="d-flex align-items-center">
                    Coming soon
                </H1>
            </PageHeader>

            <Container className="mb-3">
                <section className={classNames('row align-items-center', styles.section)}>
                    <div className="col-12 col-md-6">
                        <ProductStatusBadge status="private-beta" className="mb-4" />
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
                        <Link className={styles.link} to="/help/own" target="_blank" rel="noopener">
                            Learn more <Icon role="img" aria-label="Open in a new tab" svgPath={mdiChevronRight} />
                        </Link>
                    </div>
                    <div className="col-12 col-md-6">
                        <video
                            className="w-100 h-auto shadow percy-hide"
                            width={1280}
                            height={720}
                            autoPlay={allowAutoplay}
                            muted={true}
                            loop={true}
                            playsInline={true}
                            controls={!allowAutoplay}
                        >
                            <source
                                type="video/mp4"
                                src="https://storage.googleapis.com/sourcegraph-assets/own-search-file-owners.mp4"
                            />
                        </video>
                    </div>
                </section>
            </Container>

            <Container className="mb-3">
                <section className={classNames('row align-items-center', styles.section)}>
                    <div className="col-12 col-md-6">
                        <ProductStatusBadge status="experimental" className="mb-4" />
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
                        <Link className={styles.link} to="/help/cody" target="_blank" rel="noopener">
                            Learn more <Icon role="img" aria-label="Open in a new tab" svgPath={mdiChevronRight} />
                        </Link>
                    </div>
                    <div className="col-12 col-md-6">
                        <video
                            className="w-100 h-auto shadow percy-hide"
                            width={1280}
                            height={720}
                            autoPlay={allowAutoplay}
                            muted={true}
                            loop={true}
                            playsInline={true}
                            controls={!allowAutoplay}
                        >
                            <source
                                type="video/mp4"
                                src="https://storage.googleapis.com/sourcegraph-assets/own-search-file-owners.mp4"
                            />
                        </video>
                    </div>
                </section>
            </Container>
        </Page>
    )
}
