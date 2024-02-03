import type { FC } from 'react'

import classNames from 'classnames'

import { Text, H1, H2, H3, ButtonLink, Link } from '@sourcegraph/wildcard'

import { CodyColorIcon } from '../chat/CodyPageIcon'
import { AutocompletesIcon, ChatMessagesIcon } from '../components/CodyIcon'

import styles from './CodyDashboardPage.module.scss'

interface CodyDashboardPageProps {}

export const CodyDashboardPage: FC<CodyDashboardPageProps> = () => (
    <section className={styles.dashboardContainer}>
        <section className={styles.dashboardHero}>
            <CodyColorIcon className={styles.dashboardCodyIcon} />
            <H1 className={styles.dashboardHeroHeader}>
                Get started with <span className={styles.codyGradient}>Cody</span>
            </H1>
            <Text className={styles.dashboardHeroTagline}>
                Hey! 👋 Let’s get started with Cody — your new AI coding assistant.
            </Text>
        </section>

        <section className={styles.dashboardOnboarding}>
            <section className={styles.dashboardOnboardingIde}>
                <Text className="text-muted">Download Cody for your favorite IDE</Text>
                <Text className="text-muted">
                    Struggling with setup?{' '}
                    <Link to="" className={styles.dashboardOnboardingIdeInstallationLink}>
                        Explore installation docs.
                    </Link>
                </Text>
            </section>
            <section className={styles.dashboardOnboardingWeb}>
                <Text className="text-muted">... or try it on the web</Text>
                <ButtonLink to="/cody/chat" outline={true} className={styles.dashboardOnboardingWebLink}>
                    <CodyColorIcon className={styles.dashboardOnboardingCodyIcon} />
                    <span>Cody for web</span>
                </ButtonLink>
            </section>
        </section>

        <section className={styles.dashboardUsage}>
            <section className={styles.dashboardUsageHeader}>
                <H2>Your Usage</H2>
                <Text className={styles.dashboardUsagePlan}>Enterprise plan</Text>
            </section>

            <section className={styles.dashboardUsageDetails}>
                <section
                    className={classNames(styles.dashboardUsageDetailsGrid, styles.dashboardUsageDetailsGridFirst)}
                >
                    <section className={styles.dashboardUsageMeta}>
                        <AutocompletesIcon />
                        <span
                            className={classNames(styles.dashboardUsageMetaInfo, styles.dashboardUsageMetaInfoNumber)}
                        >
                            345
                        </span>
                        <span className={styles.dashboardUsageMetaInfo}>/</span>
                        <span className={classNames(styles.dashboardUsageMetaInfo, styles.dashboardUsageMetaInfoMax)}>
                            &#8734;
                        </span>
                    </section>
                    <H3 className={styles.dashboardUsageMetric}>Autocompletions</H3>
                    <Text className={styles.dashboardUsageTimeline}>this month</Text>
                </section>

                <section className={styles.dashboardUsageDetailsGrid}>
                    <section className={styles.dashboardUsageMeta}>
                        <ChatMessagesIcon />
                        <span
                            className={classNames(styles.dashboardUsageMetaInfo, styles.dashboardUsageMetaInfoNumber)}
                        >
                            240
                        </span>
                        <span className={styles.dashboardUsageMetaInfo}>/</span>
                        <span className={classNames(styles.dashboardUsageMetaInfo, styles.dashboardUsageMetaInfoMax)}>
                            &#8734;
                        </span>
                    </section>
                    <H3 className={styles.dashboardUsageMetric}>Chat messages</H3>
                    <Text className={styles.dashboardUsageTimeline}>this month</Text>
                </section>
            </section>
        </section>
    </section>
)
