import { FC } from 'react'
import styles from './SummaryTable.module.scss'
import { VulnerabilitiesProps } from '../../graphql/useSentinelQuery'
interface SummaryTableProps {
    vulnerabilityMatches: VulnerabilitiesProps[]
}
export const SummaryTable: FC<SummaryTableProps> = ({ vulnerabilityMatches = [] }) => {
    const totalVulnerabilities = vulnerabilityMatches.length
    const severity = getVulnerabilitySeverity(vulnerabilityMatches)
    const tableData = [
        {
            title: 'Total Vulnerabilities',
            amount: totalVulnerabilities,
        },
        {
            title: 'Critical Severity',
            amount: severity.critical ? `${severity.critical}/${totalVulnerabilities}` : '0',
        },
        {
            title: 'High Severity',
            amount: severity.high ? `${severity.high}/${totalVulnerabilities}` : '0',
        },
        {
            title: 'Medium Severity',
            amount: severity.medium ? `${severity.medium}/${totalVulnerabilities}` : '0',
        },
        {
            title: 'Repos with Vulnerabilities',
            amount: totalVulnerabilities,
        },
    ]

    return (
        <div className={styles.bar}>
            <div className={styles.container}>
                {tableData.map((data, idx) => (
                    <div key={idx} className={styles.item}>
                        <div className={styles.amount}>{data.amount}</div>
                        <div className={styles.subtitle}>{data.title}</div>
                    </div>
                ))}
            </div>
        </div>
    )
}

function getVulnerabilitySeverity(vulnerabilities: VulnerabilitiesProps[]) {
    return vulnerabilities.reduce(
        (acc, curr) => {
            if (curr.severity === 'CRITICAL') {
                acc.critical += 1
            }
            if (curr.severity === 'HIGH') {
                acc.high += 1
            }
            if (curr.severity === 'MEDIUM') {
                acc.medium += 1
            }
            return acc
        },
        { critical: 0, high: 0, medium: 0 }
    )
}
