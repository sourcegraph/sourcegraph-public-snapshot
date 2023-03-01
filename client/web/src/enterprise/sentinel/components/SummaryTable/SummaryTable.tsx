import { FC } from 'react'
import styles from './SummaryTable.module.scss'

// ! TODO: DRY it up
export const SummaryTable: FC = () => {
    return (
        <div className={styles.bar}>
            <div className={styles.container}>
                <div className={styles.item}>
                    <div className={styles.amount}>1</div>
                    <div className={styles.subtitle}>Total Vulnerabilities</div>
                </div>
                <div className={styles.item}>
                    <div className={styles.amount}>1/10</div>
                    <div className={styles.subtitle}>Critical Severity</div>
                </div>
                <div className={styles.item}>
                    <div className={styles.amount}>10/245</div>
                    <div className={styles.subtitle}>High Severity</div>
                </div>
                <div className={styles.item}>
                    <div className={styles.amount}>10/245</div>
                    <div className={styles.subtitle}>Medium Severity</div>
                </div>
                <div className={styles.item}>
                    <div className={styles.amount}>10/245</div>
                    <div className={styles.subtitle}>Repos with Vulnerabilities</div>
                </div>
            </div>
        </div>
    )
}
