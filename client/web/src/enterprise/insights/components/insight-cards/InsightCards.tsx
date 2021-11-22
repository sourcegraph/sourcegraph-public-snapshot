import classNames from 'classnames'
import React from 'react'

import { CaptureGroupInsight, PieChart, ThreeLineChart } from '../../modals/components/MediaCharts'

import styles from './InsightCards.module.scss'

export const SearchInsightCard: React.FunctionComponent = () => (
    // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
    <article tabIndex={0} className={classNames(styles.insightCard, 'card p-3')}>
        <div className={styles.insightChartContainer}>
            <ThreeLineChart viewBox="0 0 169 169" className={styles.insightChart} />
        </div>

        <div className={classNames(styles.insightCardBody, 'card-body')}>
            <h3>Based on your search query</h3>

            <p className="text-muted">
                Search-based insights let you create any data visualisation about your code based on a custom search
                query.
            </p>
        </div>
    </article>
)

export const LangStatsInsightCard: React.FunctionComponent = () => (
    // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
    <article tabIndex={0} className={classNames(styles.insightCard, 'card p-3')}>
        <div className={styles.insightChartContainer}>
            <PieChart viewBox="0 0 169 169" className={styles.insightChart} />
        </div>

        <div className={classNames(styles.insightCardBody, 'card-body')}>
            <h3>Language usage</h3>

            <p className="text-muted">Shows usage of languages in your repository based on number of lines of code.</p>
        </div>
    </article>
)

export const CaptureGroupInsightCard: React.FunctionComponent = () => (
    // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex
    <article tabIndex={0} className={classNames(styles.insightCard, 'card p-3')}>
        <div className={styles.insightChartContainer}>
            <CaptureGroupInsight className={styles.insightChart} />
        </div>

        <div className={classNames(styles.insightCardBody, 'card-body')}>
            <h3>Generated from capture groups</h3>

            <p className="text-muted">
                Data series will be generated dynamically for each unique value from the regular expression capture
                group.
            </p>
        </div>
    </article>
)
