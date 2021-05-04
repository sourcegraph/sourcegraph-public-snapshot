import { ParentSize } from '@visx/responsive'
import classnames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { LineChart } from '../../../views/ChartViewContent/charts/line/LineChart'
import { PieChart } from '../../../views/ChartViewContent/charts/pie/PieChart'

import { LINE_CHART_DATA, PIE_CHART_DATA } from './charts-mock'
import styles from './CreationIntroPage.module.scss'

/** Displays intro page for insights creation UI. */
export const CreationIntroPage: React.FunctionComponent = () => (
    <Page className="col-8">
        <PageTitle title="Create code insights" />

        <div className="mb-5">
            <h2>Create new insight</h2>

            <p className="text-muted">
                Code insights analyze your code based on any search query.{' '}
                <a
                    href="https://docs.sourcegraph.com/dev/background-information/insights"
                    target="_blank"
                    rel="noopener"
                >
                    Learn more.
                </a>
            </p>
        </div>

        <div className={classnames(styles.createIntroPageInsights, 'pb-5')}>
            <section className={classnames(styles.createIntroPageInsightCard, 'card card-body p-3')}>
                <h3>Based on your search query</h3>

                <p>
                    Search-based insights let you create any data visualization about your code based on a custom search
                    query.
                </p>

                <Link
                    to="/insights/create"
                    className={classnames(styles.createIntroPageInsightButton, 'btn', 'btn-primary')}
                >
                    Create custom insight
                </Link>

                <hr className="ml-n3 mr-n3 mt-4 mb-3" />

                <p className="text-muted">How your insight would look like:</p>
                <div className={styles.createIntroPageChartContainer}>
                    <ParentSize className={styles.createIntroPageChart}>
                        {({ width, height }) => <LineChart width={width} height={height} {...LINE_CHART_DATA} />}
                    </ParentSize>
                </div>
            </section>

            <section className={classnames(styles.createIntroPageInsightCard, 'card card-body p-3')}>
                <h3>Language usage</h3>

                <p>Shows usage of languages in your repository based on number of lines of code.</p>

                <Link
                    to="/user/settings"
                    className={classnames(styles.createIntroPageInsightButton, 'btn', 'btn-primary')}
                >
                    Set up language usage insight
                </Link>

                <hr className="ml-n3 mr-n3 mt-4 mb-3" />

                <p className="text-muted">How your insight would look like:</p>
                <div className={styles.createIntroPageChartContainer}>
                    <ParentSize className={styles.createIntroPageChart}>
                        {({ width, height }) => <PieChart width={width} height={height} {...PIE_CHART_DATA} />}
                    </ParentSize>
                </div>
            </section>

            <section className={classnames(styles.createIntroPageInsightCard, 'card card-body p-3')}>
                <h3>Based on Sourcegraph extensions</h3>

                <p>
                    Enable the extension and go to the README.md to learn how to set up code insights for selected
                    Sourcegraph extensions.
                </p>

                <Link
                    to="/extensions?query=category:Insights"
                    className={classnames(styles.createIntroPageInsightButton, 'btn', 'btn-secondary')}
                >
                    Explore the extensions
                </Link>
            </section>
        </div>
    </Page>
)
