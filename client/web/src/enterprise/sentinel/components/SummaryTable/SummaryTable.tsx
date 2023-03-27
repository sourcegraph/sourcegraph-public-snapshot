import { useQuery } from '@sourcegraph/http-client'
import { LoadingSpinner, ErrorAlert } from '@sourcegraph/wildcard'

import {
    VulnerabilityMatchesSummaryCountResult,
    VulnerabilityMatchesSummaryCountVariables,
} from '../../../../graphql-operations'
import { VULNERABILITY_MATCHES_SUMMARY_COUNT } from '../../graphql/graphqlQueries'

import styles from './SummaryTable.module.scss'

export interface VulnerabilitiesProps {
    sourceID: string
    details: string
    summary: string
    published: string
    modified: string
    cvssScore: string
    severity: string
}
export function SummaryTable(): JSX.Element {
    const { data, error, loading } = useQuery<
        VulnerabilityMatchesSummaryCountResult,
        VulnerabilityMatchesSummaryCountVariables
    >(VULNERABILITY_MATCHES_SUMMARY_COUNT, { fetchPolicy: 'cache-and-network' })

    if (!data || loading) {
        return <LoadingSpinner />
    }

    if (error) {
        return <ErrorAlert error={new Error('Sentinel summary is not available')} />
    }

    const summary = getSummary(data)
    const tableData = [
        {
            title: 'Total Vulnerabilities',
            amount: summary.total,
        },
        {
            title: 'Critical Severity',
            amount: summary.critical ? `${summary.critical}/${summary.total}` : '0',
        },
        {
            title: 'High Severity',
            amount: summary.high ? `${summary.high}/${summary.total}` : '0',
        },
        {
            title: 'Medium Severity',
            amount: summary.medium ? `${summary.medium}/${summary.total}` : '0',
        },
        {
            title: 'Repos with Vulnerabilities',
            amount: summary.repository,
        },
    ]

    return (
        <div className={styles.bar}>
            <div className={styles.container}>
                {tableData.map(data => (
                    <div key={data.title} className={styles.item}>
                        <div className={styles.amount}>{data.amount}</div>
                        <div className={styles.subtitle}>{data.title}</div>
                    </div>
                ))}
            </div>
        </div>
    )
}

interface VulnerabilityMatchesSummaryCount {
    critical: number
    high: number
    medium: number
    low: number
    total: number
    repository: number
}

declare const vulnerabilityMatchesSummaryCounts: {
    vulnerabilityMatchesSummaryCounts: {
        critical: number
        high: number
        medium: number
        low: number
        repository: number
    }
}

function getSummary(summary: typeof vulnerabilityMatchesSummaryCounts): VulnerabilityMatchesSummaryCount {
    const {
        critical = 0,
        high = 0,
        medium = 0,
        low = 0,
        repository = 0,
    } = summary?.vulnerabilityMatchesSummaryCounts || {}

    const total = high + medium + low + critical
    return {
        critical,
        high,
        medium,
        low,
        total,
        repository,
    }
}
