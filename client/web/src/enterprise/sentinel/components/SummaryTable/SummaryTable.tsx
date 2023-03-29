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

    const tableData = getSummary(data)

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

declare const vulnerabilityMatchesSummaryCounts: {
    vulnerabilityMatchesSummaryCounts: {
        critical: number
        high: number
        medium: number
        low: number
        repository: number
    }
}

interface SummaryData {
    title: string
    amount: string | number
}

function getSummary(summary: typeof vulnerabilityMatchesSummaryCounts): SummaryData[] {
    const {
        critical = 0,
        high = 0,
        medium = 0,
        low = 0,
        repository = 0,
    } = summary?.vulnerabilityMatchesSummaryCounts || {}

    const total = high + medium + low + critical
    return [
        {
            title: 'Total Vulnerabilities',
            amount: total,
        },
        {
            title: 'Critical Severity',
            amount: critical ? `${critical}/${total}` : '0',
        },
        {
            title: 'High Severity',
            amount: high ? `${high}/${total}` : '0',
        },
        {
            title: 'Medium Severity',
            amount: medium ? `${medium}/${total}` : '0',
        },
        {
            title: 'Repos with Vulnerabilities',
            amount: repository,
        },
    ]
}
