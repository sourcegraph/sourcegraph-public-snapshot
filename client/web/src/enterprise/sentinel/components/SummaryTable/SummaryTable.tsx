import { FC } from 'react'
import styles from './SummaryTable.module.scss'
import { VulnerabilitiesProps } from '../../graphql/useSentinelQuery'
interface SummaryTableProps {
    vulnerabilityMatches: VulnerabilitiesProps[]
}
// ! TODO: DRY it up
export const SummaryTable: FC<SummaryTableProps> = ({ vulnerabilityMatches = [] }) => {
    const totalVulnerabilities = vulnerabilityMatches.length
    const severity = vulnerabilityMatches.reduce(
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

    return (
        <div className={styles.bar}>
            <div className={styles.container}>
                <div className={styles.item}>
                    <div className={styles.amount}>{totalVulnerabilities}</div>
                    <div className={styles.subtitle}>Total Vulnerabilities</div>
                </div>
                <div className={styles.item}>
                    <div className={styles.amount}>
                        {severity.critical ? `${severity.critical}/${totalVulnerabilities}` : '0'}
                    </div>
                    <div className={styles.subtitle}>Critical Severity</div>
                </div>
                <div className={styles.item}>
                    <div className={styles.amount}>
                        {severity.high ? `${severity.high}/${totalVulnerabilities}` : '0'}
                    </div>
                    <div className={styles.subtitle}>High Severity</div>
                </div>
                <div className={styles.item}>
                    <div className={styles.amount}>
                        {severity.medium ? `${severity.medium}/${totalVulnerabilities}` : '0'}
                    </div>
                    <div className={styles.subtitle}>Medium Severity</div>
                </div>
                <div className={styles.item}>
                    <div className={styles.amount}>{totalVulnerabilities}</div>
                    <div className={styles.subtitle}>Repos with Vulnerabilities</div>
                </div>
            </div>
        </div>
    )
}
