import classNames from 'classnames'

import styles from './DataSummary.module.scss'

export interface DataSummaryItem {
    value: number
    label: string

    className?: string
    valueClassName?: string
}

interface DataSummaryProps {
    items: DataSummaryItem[]
    className?: string
}

export const DataSummary: React.FunctionComponent<DataSummaryProps> = ({ items, className }) => (
    <div className={classNames(styles.summary, className)}>
        {items.map(({ label, value, className, valueClassName }, index) => (
            <div className={classNames(styles.summaryItem, className)} key={index}>
                <div className={classNames(styles.summaryNumber, valueClassName)}>{value}</div>
                <div className="text-muted">{label}</div>
            </div>
        ))}
    </div>
)
