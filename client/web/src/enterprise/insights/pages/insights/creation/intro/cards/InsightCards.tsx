import React from 'react'

import classNames from 'classnames'

import { Link, Button, CardBody, Card, H3, H2 } from '@sourcegraph/wildcard'

import {
    CaptureGroupInsightChart,
    LangStatsInsightChart,
    SearchBasedInsightChart,
} from '../../../../../modals/components/MediaCharts'

import styles from './InsightCards.module.scss'

interface InsightCardProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {}

/**
 * Low-level styled component for building insight link card for
 * the creation page gallery.
 */
const InsightCard: React.FunctionComponent<InsightCardProps> = props => {
    const { children, ...otherProps } = props

    return (
        <Card
            as="button"
            {...otherProps}
            type="button"
            className={classNames(styles.card, 'p-3', otherProps.className)}
        >
            {children}

            <Button as="div" className="mt-3 w-100" variant="secondary" size="sm">
                Create
            </Button>
        </Card>
    )
}

interface InsightCardBodyProps {
    title: string
    className?: string
}

const InsightCardBody: React.FunctionComponent<InsightCardBodyProps> = props => {
    const { title, className, children } = props

    return (
        <CardBody className={classNames(styles.cardBody, className, 'flex-1')}>
            <H3 as={H2} className={styles.cardTitle}>
                {title}
            </H3>
            <p className="d-flex flex-column text-muted m-0">{children}</p>
        </CardBody>
    )
}

const InsightCardExampleBlock: React.FunctionComponent = props => (
    <footer className={styles.cardFooter}>
        <small className="text-muted">Example use</small>
        <small className={styles.cardExampleBlock}>{props.children}</small>
    </footer>
)

export const SearchInsightCard: React.FunctionComponent<InsightCardProps> = props => (
    <InsightCard {...props}>
        <SearchBasedInsightChart className={styles.chart} />
        <InsightCardBody title="Track changes" className="mb-3">
            Insight <b>based on a custom Sourcegraph search query</b> that creates visualization of the data series you
            will define <b>manually.</b>
        </InsightCardBody>

        <InsightCardExampleBlock>Tracking architecture, naming, or language migrations.</InsightCardExampleBlock>
    </InsightCard>
)

export const LangStatsInsightCard: React.FunctionComponent<InsightCardProps> = props => (
    <InsightCard {...props}>
        <LangStatsInsightChart viewBox="0 0 169 148" className={styles.chart} />
        <InsightCardBody title="Language usage">
            Shows usage of languages in your repository based on number of lines of code.
        </InsightCardBody>
    </InsightCard>
)

export const CaptureGroupInsightCard: React.FunctionComponent<InsightCardProps> = props => (
    <InsightCard {...props}>
        <CaptureGroupInsightChart className={styles.chart} />

        <InsightCardBody title="Detect and track patterns" className="mb-3">
            Data series will be generated dynamically for each unique value from the
            <b> regular expression capture group </b> included in the search query. Chart will be updated as new values
            appear in the code base.
        </InsightCardBody>

        <InsightCardExampleBlock>Detecting and tracking language or package versions.</InsightCardExampleBlock>
    </InsightCard>
)

// a11y-ignore
// Rule: nested-interactive (Interactive controls must not be nested)
// InsightCard is a button and it contains Link, it causes this issue,
// we should consider keeping link + make InsightCard a div or make InsightCard a Link + make Link non-interactive
export const ExtensionInsightsCard: React.FunctionComponent<InsightCardProps> = props => (
    <InsightCard {...props} className={classNames(styles.cardExtensionCard, 'a11y-ignore')}>
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

        <InsightCardBody title="Based on Sourcegraph extensions">
            Enable the extension and go to the README.md to learn how to set up code insights for selected Sourcegraph
            extensions.
            <Link to="/extensions?query=category:Insights&experimental=true">Explore the extensions</Link>
        </InsightCardBody>
    </InsightCard>
)
