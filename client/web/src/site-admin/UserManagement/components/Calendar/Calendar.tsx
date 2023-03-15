import classNames from 'classnames'
import ReactCalendar from 'react-calendar'

import { Container } from '@sourcegraph/wildcard'

import styles from './Calendar.module.scss'

interface CalendarDateProps {
    isRange?: false
    value?: Date | null
    onChange: (value: Date) => void
}

interface CalendarDateRangeProps {
    isRange: true
    value?: [Date | null, Date | null] | null
    onChange: (value: [Date, Date]) => void
}

type CalendarProps = {
    className?: string
    maxDate?: Date
    minDate?: Date
    highlightToday?: boolean
} & (CalendarDateRangeProps | CalendarDateProps)

/**
 * Renders a calendar component which supports single date or range selection.
 *
 * **NOTE:** This is an `EXPERIMENTAL` component and built on top of `react-calendar` package with Sourcegraph CSS styling on top.
 * It intentionally, omits other `react-calendar` props/features to not over-complicate and use as simple calendar, in case if we migrate to another calendar library or build our own.
 *
 * Depending on `isRange` value `true | false`, the component will render a single or range selection calendar as well as infer correct TS props for the range selection.
 *
 */
export const Calendar: React.FunctionComponent<CalendarProps> = ({
    className,
    value,
    onChange,
    isRange,
    minDate,
    maxDate,
    highlightToday,
}) => (
    <Container className={classNames(styles.container, highlightToday && styles.highlightToday, className)}>
        <ReactCalendar
            /**
             * The underlying react-calendar component "onChange" handler arguments differ depending on the "selectRange"
             * however type-wise is not differentiated thus TS cannot infer the correct type.
             */
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
