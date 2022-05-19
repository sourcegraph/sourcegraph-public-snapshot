import React from 'react'

import classNames from 'classnames'

import { Link, Button, CardBody, Card, Typography } from '@sourcegraph/wildcard'

import {
    CaptureGroupInsightChart,
    LangStatsInsightChart,
    SearchBasedInsightChart,
} from '../../../../../modals/components/MediaCharts'

import styles from './InsightCards.module.scss'

interface InsightCardProps extends React.HTMLAttributes<HTMLDivElement> {
    handleCreate?: () => void
}

/**
 * Low-level styled component for building insight link card for
 * the creation page gallery.
 */
const InsightCard: React.FunctionComponent<React.PropsWithChildren<InsightCardProps>> = props => {
    const { children, onClick, handleCreate, ...otherProps } = props

    return (
        <Card {...otherProps} className={classNames(styles.card, 'p-3', otherProps.className)}>
            {children}

            <Button className="mt-3 w-100" variant="secondary" onClick={handleCreate}>
                Create
            </Button>
        </Card>
    )
}

interface InsightCardBodyProps {
    title: string
    className?: string
}

const InsightCardBody: React.FunctionComponent<React.PropsWithChildren<InsightCardBodyProps>> = props => {
    const { title, className, children } = props

    return (
        <CardBody className={classNames(styles.cardBody, className, 'flex-1')}>
            <Typography.H3 as={Typography.H2} className={styles.cardTitle}>
                {title}
            </Typography.H3>
            <p className="d-flex flex-column text-muted m-0">{children}</p>
        </CardBody>
    )
}

const InsightCardExampleBlock: React.FunctionComponent<React.PropsWithChildren<unknown>> = props => (
    <footer className={styles.cardFooter}>
        <small className="text-muted">Example use</small>
        <small className={styles.cardExampleBlock}>{props.children}</small>
    </footer>
)

export const SearchInsightCard: React.FunctionComponent<React.PropsWithChildren<InsightCardProps>> = props => (
    <InsightCard {...props}>
        <SearchBasedInsightChart className={styles.chart} />
        <InsightCardBody title="Track changes" className="mb-3">
            Insight <b>based on a custom Sourcegraph search query</b> that creates visualization of the data series you
            will define <b>manually.</b>
        </InsightCardBody>

        <InsightCardExampleBlock>Tracking architecture, naming, or language migrations.</InsightCardExampleBlock>
    </InsightCard>
)

export const LangStatsInsightCard: React.FunctionComponent<React.PropsWithChildren<InsightCardProps>> = props => (
    <InsightCard {...props}>
        <LangStatsInsightChart viewBox="0 0 169 148" className={styles.chart} />
        <InsightCardBody title="Language usage">
            Shows usage of languages in your repository based on number of lines of code.
        </InsightCardBody>
    </InsightCard>
)

export const CaptureGroupInsightCard: React.FunctionComponent<React.PropsWithChildren<InsightCardProps>> = props => (
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

export const ExtensionInsightsCard: React.FunctionComponent<React.PropsWithChildren<InsightCardProps>> = props => (
    <InsightCard {...props} className={styles.cardExtensionCard}>
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
