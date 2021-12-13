import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import { CaptureGroupInsight, PieChart, ThreeLineChart } from '../../../../../modals/components/MediaCharts'

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
        <ThreeLineChart viewBox="0 0 169 148" className={styles.chart} />
        <CardBody title="Track">
            Insight <b>based on a custom Sourcegraph search query</b> that creates visualization of the data series you
            will define <b>manually.</b>
        </CardBody>
    </Card>
)

export const LangStatsInsightCard: React.FunctionComponent<CardProps> = props => (
    <Card {...props}>
        <PieChart viewBox="0 0 169 148" className={styles.chart} />
        <CardBody title="Language usage">
            Shows usage of languages in your repository based on number of lines of code.
        </CardBody>
    </Card>
)

export const CaptureGroupInsightCard: React.FunctionComponent<CardProps> = props => (
    <Card {...props} footerText="Detecting and tracking language or package versions.">
        <div className={styles.captureChartWrapper}>
            <CaptureGroupInsight className={styles.chart} />
            <CaptureGroupIcon className={styles.captureChartIcon} />
        </div>

        <CardBody title="Detect and track">
            Data series will be generated dynamically for each unique value from the
            <b> regular expression capture group </b> included in the search query. Chart will be updated as new values
            appear in the code base.
        </CardBody>
    </Card>
)

const CaptureGroupIcon: React.FunctionComponent<React.HTMLAttributes<HTMLElement>> = props => (
    <div className={classNames(props.className, styles.captureGroupIcon)}>
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path
                d="M9.76238 9.82535C9.49591 9.86573 9.22943 9.88995 8.95488 9.88995C8.68033 9.88995 8.41386 9.86573 8.14738 9.82535V6.99103L6.12863 8.99363C5.72488 8.6787 5.32113 8.27495 5.00621 7.8712L7.00881 5.85245H4.17448C4.13411 5.58598 4.10988 5.3195 4.10988 5.04495C4.10988 4.7704 4.13411 4.50393 4.17448 4.23745H7.00881L5.00621 2.2187C5.15963 2.01683 5.32113 1.81495 5.53108 1.62115C5.72488 1.4112 5.92676 1.2497 6.12863 1.09628L8.14738 3.09888V0.264551C8.41386 0.224176 8.68033 0.199951 8.95488 0.199951C9.22943 0.199951 9.49591 0.224176 9.76238 0.264551V3.09888L11.7811 1.09628C12.1849 1.4112 12.5886 1.81495 12.9036 2.2187L10.901 4.23745H13.7353C13.7757 4.50393 13.7999 4.7704 13.7999 5.04495C13.7999 5.3195 13.7757 5.58598 13.7353 5.85245H10.901L12.9036 7.8712C12.7501 8.07308 12.5886 8.27495 12.3787 8.46875C12.1849 8.6787 11.983 8.8402 11.7811 8.99363L9.76238 6.99103V9.82535ZM0.879883 11.505C0.879883 11.0766 1.05003 10.6658 1.35291 10.363C1.65578 10.0601 2.06656 9.88995 2.49488 9.88995C2.92321 9.88995 3.33399 10.0601 3.63686 10.363C3.93973 10.6658 4.10988 11.0766 4.10988 11.505C4.10988 11.9333 3.93973 12.3441 3.63686 12.6469C3.33399 12.9498 2.92321 13.12 2.49488 13.12C2.06656 13.12 1.65578 12.9498 1.35291 12.6469C1.05003 12.3441 0.879883 11.9333 0.879883 11.505Z"
                fill="var(--body-color)"
            />
        </svg>
    </div>
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
