import classNames from 'classnames'

import { LinkOrSpan } from '@sourcegraph/wildcard'

import styles from './DataSummary.module.scss'

export interface DataSummaryItem {
    value: JSX.Element
    label: string
    to?: string

    className?: string
    valueClassName?: string
}

interface DataSummaryProps {
    items: DataSummaryItem[]
    className?: string
}

export const DataSummary: React.FunctionComponent<DataSummaryProps> = ({ items, className }) => (
    <div className={classNames(styles.summary, className)}>
        {items.map(({ label, value, className, valueClassName, to }, index) => (
            <LinkOrSpan to={to} className={classNames(styles.summaryItem, className)} key={index}>
                <div className={classNames(styles.summaryNumber, valueClassName)}>{value}</div>
                <div className={styles.summaryLabel}>{label}</div>
            </LinkOrSpan>
        ))}
    </div>
)
