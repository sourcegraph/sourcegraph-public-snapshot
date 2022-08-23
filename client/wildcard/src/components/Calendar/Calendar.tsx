import classNames from 'classnames'
import ReactCalendar from 'react-calendar'

import { Container } from '@sourcegraph/wildcard'

import styles from './Calendar.module.scss'

type CalendarProps = {
    className?: string
    maxDate?: Date
    minDate?: Date
    highlightToday?: boolean
} & (
    | {
          isRange: true
          value: [Date | null, Date | null]
          onChange: (value: [Date, Date]) => void
      }
    | {
          isRange?: false
          value: Date | null | undefined
          onChange: (value: Date) => void
      }
)

export function Calendar({
    className,
    value,
    onChange,
    isRange,
    minDate,
    maxDate,
    highlightToday,
}: CalendarProps): JSX.Element {
    return (
        <Container className={classNames(styles.container, styles.highlightToday, className)}>
            <ReactCalendar
                // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                // @ts-ignore
                onChange={onChange}
                value={value}
                selectRange={isRange}
                view="month"
                maxDate={maxDate}
                minDate={minDate}
                showFixedNumberOfWeeks={true}
            />
        </Container>
    )
}
