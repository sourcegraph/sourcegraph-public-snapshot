import classNames from 'classnames'
import React from 'react'
import { Link } from 'react-router-dom'

import {
    CaptureGroupInsightChart,
    LangStatsInsightChart,
    SearchBasedInsightChart,
} from '../../../../../modals/components/MediaCharts'

import styles from './InsightCards.module.scss'

interface CardProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {}

/**
 * Low-level styled component for building insight link card for
 * the creation page gallery.
 */
const Card: React.FunctionComponent<CardProps> = props => {
    const { children, ...otherProps } = props

    return (
        <button {...otherProps} type="button" className={classNames(styles.card, 'card p-3', otherProps.className)}>
            {children}

            <div className="btn btn-sm btn-secondary mt-3 w-100">Create</div>
        </button>
    )
}

interface CardBodyProps {
    title: string
    className?: string
}

const CardBody: React.FunctionComponent<CardBodyProps> = props => {
    const { title, className, children } = props

    return (
        <div className={classNames(styles.cardBody, className, 'card-body flex-1')}>
            <h3 className={styles.cardTitle}>{title}</h3>

            <p className="d-flex flex-column text-muted m-0">{children}</p>
        </div>
    )
}

const CardExampleBlock: React.FunctionComponent = props => (
    <footer className={styles.cardFooter}>
        <small className="text-muted">Example use</small>
        <small className={styles.cardExampleBlock}>{props.children}</small>
    </footer>
)

export const SearchInsightCard: React.FunctionComponent<CardProps> = props => (
    <Card {...props}>
        <SearchBasedInsightChart className={styles.chart} />
        <CardBody title="Track changes" className="mb-3">
            Insight <b>based on a custom Sourcegraph search query</b> that creates visualization of the data series you
            will define <b>manually.</b>
        </CardBody>

        <CardExampleBlock>Tracking architecture, naming, or language migrations.</CardExampleBlock>
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
    <Card {...props}>
        <CaptureGroupInsightChart className={styles.chart} />

        <CardBody title="Detect and track patterns" className="mb-3">
            Data series will be generated dynamically for each unique value from the
            <b> regular expression capture group </b> included in the search query. Chart will be updated as new values
            appear in the code base.
        </CardBody>

        <CardExampleBlock>Detecting and tracking language or package versions.</CardExampleBlock>
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
