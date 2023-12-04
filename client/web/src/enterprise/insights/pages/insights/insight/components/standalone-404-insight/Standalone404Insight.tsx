import type { FunctionComponent } from 'react'

import { Button, Link, H2, Text } from '@sourcegraph/wildcard'

import styles from './Standalone404Insight.module.scss'

export const Standalone404Insight: FunctionComponent = () => (
    <div className={styles.container}>
        <GraphicInsightChart className={styles.chart} />

        <H2 className="mb-3">Insight not found</H2>
        <Text>Insight may not exist or you may not have permission to view it.</Text>

        <Button as={Link} to="/insights/all" variant="primary" className={styles.redirectButton}>
            Go to 'All insights'
        </Button>
    </div>
)

const GraphicInsightChart: FunctionComponent<{ className: string }> = ({ className }) => (
    <svg width="263" height="134" fill="none" xmlns="http://www.w3.org/2000/svg" className={className}>
        <path
            d="M6.715 129.3H90.01l41.91-42.957 42.957-34.052 41.909-19.383 41.386-23.05"
            stroke="var(--static-chart-blue)"
            strokeWidth="3"
            strokeLinecap="round"
            strokeLinejoin="round"
        />

        <path
            d="M6.19 46.529 48.1 5.143h42.434l42.433 3.143 41.909 100.583 40.862 4.715h41.909"
            stroke="var(--static-chart-purple)"
            strokeWidth="3"
            strokeLinecap="round"
            strokeLinejoin="round"
        />

        <path
            d="M6.19 7.763 48.1 41.29l41.91-3.143 42.433 3.143 42.433 19.907 41.386 3.143 41.385 2.096"
            stroke="var(--static-chart-orange)"
            strokeWidth="3"
            strokeLinecap="round"
            strokeLinejoin="round"
        />

        <circle
            cx="48.624"
            cy="5.667"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-purple)"
            strokeWidth="2"
        />
        <circle
            cx="90.533"
            cy="4.619"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-purple)"
            strokeWidth="2"
        />
        <circle
            cx="131.395"
            cy="8.81"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-purple)"
            strokeWidth="2"
        />
        <circle
            cx="5.667"
            cy="47.577"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-purple)"
            strokeWidth="2"
        />
        <circle
            cx="214.166"
            cy="113.583"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-purple)"
            strokeWidth="2"
        />
        <circle
            cx="258.171"
            cy="113.583"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-purple)"
            strokeWidth="2"
        />
        <circle
            cx="174.352"
            cy="108.345"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-purple)"
            strokeWidth="2"
        />

        <circle
            cx="4.619"
            cy="7.763"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-orange)"
            strokeWidth="2"
        />
        <circle
            cx="48.624"
            cy="42.338"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-orange)"
            strokeWidth="2"
        />
        <circle
            cx="92.629"
            cy="39.194"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-orange)"
            strokeWidth="2"
        />
        <circle
            cx="131.395"
            cy="41.29"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-orange)"
            strokeWidth="2"
        />
        <circle
            cx="177.495"
            cy="62.245"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-orange)"
            strokeWidth="2"
        />
        <circle
            cx="216.262"
            cy="64.34"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-orange)"
            strokeWidth="2"
        />
        <circle
            cx="256.076"
            cy="65.388"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-orange)"
            strokeWidth="2"
        />

        <circle
            cx="174.352"
            cy="52.815"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-blue)"
            strokeWidth="2"
        />
        <circle
            cx="132.443"
            cy="85.295"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-blue)"
            strokeWidth="2"
        />
        <circle
            cx="89.486"
            cy="129.299"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-blue)"
            strokeWidth="2"
        />
        <circle
            cx="6.715"
            cy="129.299"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-blue)"
            strokeWidth="2"
        />
        <circle
            cx="258.171"
            cy="10.906"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-blue)"
            strokeWidth="2"
        />
        <circle
            cx="219.405"
            cy="30.813"
            r="3.619"
            fill="var(--body-bg)"
            stroke="var(--static-chart-blue)"
            strokeWidth="2"
        />
    </svg>
)
