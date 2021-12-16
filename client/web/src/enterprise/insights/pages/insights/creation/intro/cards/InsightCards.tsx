import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import {
    CaptureGroupInsightChart,
    LangStatsInsightChart,
    SearchBasedInsightChart,
} from '../../../../../modals/components/MediaCharts'

import styles from './InsightCards.module.scss'

interface CardProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    footerText?: string
}

/**
 * Low-level styled component for building insight link card for
 * the creation page gallery.
 */
const Card: React.FunctionComponent<CardProps> = props => {
    const { children, footerText, ...otherProps } = props

    return (
        <button {...otherProps} type="button" className={classNames(styles.card, 'card p-3', otherProps.className)}>
            {children}

            {footerText && (
                <footer className="d-flex flex-column mt-3">
                    <small className="text-muted">Example use</small>
                    <span>{footerText}</span>
                </footer>
            )}
        </button>
    )
}

const CardBody: React.FunctionComponent<{ title: string }> = props => {
    const { title, children } = props

    return (
        <div className={classNames(styles.cardBody, 'card-body flex-1')}>
            <h3 className="mb-3">{title}</h3>

            <p className="d-flex flex-column text-muted m-0">{children}</p>
        </div>
    )
}

export const SearchInsightCard: React.FunctionComponent<CardProps> = props => (
    <Card {...props} footerText="Redis, PostgreSQL and SQLite database usage.">
        <SearchBasedInsightChart className={styles.chart} />
        <CardBody title="Track">
            Insight <b>based on a custom Sourcegraph search query</b> that creates visualization of the data series you
            will define <b>manually.</b>
        </CardBody>
    </Card>
)

export const LangStatsInsightCard: React.FunctionComponent<CardProps> = props => (
    <Card {...props}>
        <LangStatsInsightChart viewBox="0 0 169 148" className={styles.chart} />
        <CardBody title="Language usage">
            Shows usage of languages in your repository based on number of lines of code.
        </CardBody>
    </Card>
)

export const CaptureGroupInsightCard: React.FunctionComponent<CardProps> = props => (
    <Card {...props} footerText="Detecting and tracking language or package versions.">
        <CaptureGroupInsightChart className={styles.chart} />

        <CardBody title="Detect and track">
            Data series will be generated dynamically for each unique value from the
            <b> regular expression capture group </b> included in the search query. Chart will be updated as new values
            appear in the code base.
        </CardBody>
    </Card>
)

export const ExtensionInsightsCard: React.FunctionComponent<CardProps> = props => (
    <Card {...props} className={styles.cardExtensionCard}>
        <div className={styles.images}>
            <img
                className={styles.image}
                src={`${window.context?.assetsRoot || ''}/img/codecov.png`}
                data-skip-percy={true}
                alt="Codecov logo"
            />
            <img
                className={styles.image}
                src={`${window.context?.assetsRoot || ''}/img/eslint.png`}
                data-skip-percy={true}
                alt="Eslint logo"
            />
            <img
                className={styles.image}
                src={`${window.context?.assetsRoot || ''}/img/snyk.png`}
                data-skip-percy={true}
                alt="Snyk logo"
            />
        </div>

        <CardBody title="Based on Sourcegraph extensions">
            Enable the extension and go to the README.md to learn how to set up code insights for selected Sourcegraph
            extensions. <Link to="/extensions?query=category:Insights&experimental=true">Explore the extensions</Link>
        </CardBody>
    </Card>
)
