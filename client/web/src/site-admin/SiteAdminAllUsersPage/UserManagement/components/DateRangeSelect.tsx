import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { mdiChevronDown, mdiClose } from '@mdi/js'
import classNames from 'classnames'
import {
    endOfMonth,
    endOfToday,
    endOfWeek,
    startOfMonth,
    startOfToday,
    startOfWeek,
    format,
    endOfDay,
    subMonths,
} from 'date-fns'

import {
    Button,
    Container,
    Icon,
    Popover,
    PopoverContent,
    PopoverOpenEvent,
    PopoverTrigger,
    Checkbox,
    Tooltip,
} from '@sourcegraph/wildcard'

import { Calendar } from './Calendar'

import styles from './DateRangeSelect.module.scss'

export interface DateRangeSelectProps {
    placeholder?: string
    value?: [Date, Date]
    isRequired?: boolean
    onChange?: (value?: [Date, Date], isNegated?: boolean) => void
    defaultIsOpen?: boolean
    // If provided will render additional checkbox and pass value to onChange prop
    negation?: {
        label: string
        value?: boolean
        message?: React.ReactNode
    }
    className?: string
}
export const DateRangeSelect: React.FunctionComponent<DateRangeSelectProps> = ({
    placeholder = 'Select a date range',
    value,
    onChange,
    isRequired,
    negation,
    className,
    defaultIsOpen = false,
}) => {
    const predefinedDates: [Date, Date, string][] = useMemo(() => {
        const now = new Date()

        return [
            [startOfToday(), endOfToday(), 'Today'],
            [startOfWeek(now), endOfWeek(now), 'This week'],
            [startOfMonth(now), endOfMonth(now), 'This month'],
            [subMonths(now, 3), endOfDay(now), 'Last 3 months'],
        ]
    }, [])

    const [range, setRange] = useState<[Date, Date] | undefined>(value)
    const [isOpen, setIsOpen] = useState(defaultIsOpen)
    const [isNegated, setIsNegated] = useState<boolean | undefined>(negation?.value ?? false)

    const handleCancel = useCallback((): void => {
        setIsOpen(false)
    }, [])

    const handleOpenChange = useCallback((event: PopoverOpenEvent): void => {
        setIsOpen(event.isOpen)
    }, [])

    useEffect(() => {
        if (!isOpen) {
            setRange(value)
            setIsNegated(negation?.value ?? false)
        }
    }, [isOpen, negation?.value, value])

    const handleApply = useCallback((): void => {
        onChange?.(range, isNegated)
        setIsOpen(false)
    }, [isNegated, onChange, range])

    const handleClear = useCallback(() => {
        setRange(() => undefined)
        setIsNegated(false)
    }, [])

    const { tooltip, label } = useMemo(() => {
        let tooltipText = ''
        let labelText = ''
        if (negation?.value) {
            tooltipText += 'Not in '
            labelText = '-:'
        }
        if (value) {
            tooltipText += ` ${format(value[0], 'MMM d, yyyy')} - ${format(value[1], 'MMM d, yyyy')}`
            labelText += `${format(value[0], 'M/d')}-${format(value[1], 'M/d')}`
        } else if (negation?.value) {
            tooltipText += 'all available time'
            labelText += 'all time'
        } else {
            labelText = placeholder
        }

        return { label: labelText, tooltip: tooltipText }
    }, [negation?.value, placeholder, value])

    return (
        <Popover isOpen={isOpen} onOpenChange={handleOpenChange}>
            <PopoverTrigger as={Button} className={className} display="block" variant="secondary" outline={true}>
                <div className="d-flex align-items-center justify-content-between">
                    <Tooltip content={tooltip}>
                        <span>{label}</span>
                    </Tooltip>
                    <Icon aria-label="Arrow down" svgPath={mdiChevronDown} className="ml-1" />
                </div>
            </PopoverTrigger>
            <PopoverContent focusLocked={false}>
                <Container className="d-flex flex-column">
                    {negation && (
                        <Checkbox
                            name={negation.label}
                            id={negation.label}
                            wrapperClassName={classNames('mb-3', styles.negationCheckbox)}
                            checked={isNegated}
                            onChange={event => setIsNegated(event.target.checked)}
                            label={negation.label}
                            message={negation.message}
                        />
                    )}
                    <div className="d-flex justify-content-start mb-2">
                        {predefinedDates.map(([start, end, label]) => (
                            <Button
                                key={label}
                                size="sm"
                                variant="secondary"
                                className="mr-2"
                                outline={true}
                                onClick={() => setRange([start, end])}
                            >
                                {label}
                            </Button>
                        ))}
                    </div>
                    <Calendar
                        highlightToday={true}
                        className="border-0 p-0 m-0"
                        isRange={true}
                        value={range}
                        onChange={setRange}
                    />
                    <div className="d-flex justify-content-between align-items-start mt-3">
                        <div>
                            <Button
                                size="sm"
                                variant="secondary"
                                className="mr-2 align-self-start"
                                outline={true}
                                onClick={handleClear}
                            >
                                <Icon aria-hidden={true} className="mr-1" svgPath={mdiClose} />
                                Clear
                            </Button>
                        </div>
                        <div>
                            <Button size="sm" className="mr-2" variant="secondary" onClick={handleCancel}>
                                Cancel
                            </Button>
                            <Button
                                size="sm"
                                variant="primary"
                                onClick={handleApply}
                                disabled={isRequired && !range && isNegated}
                            >
                                Apply
                            </Button>
                        </div>
                    </div>
                </Container>
            </PopoverContent>
        </Popover>
    )
}
