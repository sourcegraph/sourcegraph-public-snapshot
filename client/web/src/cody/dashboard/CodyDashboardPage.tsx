import type { FC } from 'react'

import { Text, H2 } from '@sourcegraph/wildcard'

import styles from './CodyDashboardPage.module.scss'

interface CodyDashboardPageProps {}

export const CodyDashboardPage: FC<CodyDashboardPageProps> = () => (
    <section className={styles.dashboardContainer}>
        <section className={styles.dashboardUsage}>
            <section className={styles.dashboardUsageHeader}>
                <H2>Your Usage</H2>
                <Text className={styles.dashbordUsagePlan}>Enterprise plan</Text>
            </section>
        </section>
    </section>
)
