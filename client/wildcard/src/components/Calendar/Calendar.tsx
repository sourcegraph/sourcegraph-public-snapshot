import classNames from 'classnames'
import ReactCalendar from 'react-calendar'

import { Container } from '@sourcegraph/wildcard'

import styles from './Calendar.module.scss'

interface CalendarProps {
    mode?: 'range' | 'single'
    className?: string
    value: Date | null | undefined | [Date | null, Date | null]
    onChange: (value: Date | ([Date] | [Date, Date])) => void
}

export function Calendar({ className, value, onChange, mode }: CalendarProps): JSX.Element {
    return (
        <Container className={classNames(styles.container, className)}>
            <ReactCalendar onChange={onChange} value={value} selectRange={mode === 'range'} view="month" />
        </Container>
    )
}
