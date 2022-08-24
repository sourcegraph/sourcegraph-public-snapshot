import React, { useCallback, useMemo, useState } from 'react'

import { mdiChevronDown, mdiClose } from '@mdi/js'
import { endOfMonth, endOfToday, endOfWeek, startOfMonth, startOfToday, startOfWeek, format } from 'date-fns'

import {
    Button,
    Calendar,
    Container,
    Icon,
    Popover,
    PopoverContent,
    PopoverOpenEvent,
    PopoverTrigger,
    Tooltip,
} from '@sourcegraph/wildcard'

interface DateRangeSelectProps {
    placeholder?: string
    value?: [Date, Date] | null
    onChange?: (value?: [Date, Date] | null) => void
    nullLabel?: string
}
export const DateRangeSelect: React.FunctionComponent<DateRangeSelectProps> = ({
    placeholder = 'Select a date range',
    value,
    onChange,
    nullLabel,
}) => {
    const predefinedDates: [Date, Date, string][] = useMemo(() => {
        const now = new Date()

        return [
            [startOfToday(), endOfToday(), 'Today'],
            [startOfWeek(now), endOfWeek(now), 'This week'],
            [startOfMonth(now), endOfMonth(now), 'This month'],
        ]
    }, [])

    const [isOpen, setIsOpen] = useState(false)
    const handleOpenChange = useCallback((event: PopoverOpenEvent): void => {
        setIsOpen(event.isOpen)
    }, [])

    const handleChange = (value?: [Date, Date] | null): void => {
        onChange?.(value)
        setIsOpen(false)
    }

    return (
        <Popover isOpen={isOpen} onOpenChange={handleOpenChange}>
            <PopoverTrigger as={Button} display="block" variant="secondary" outline={true}>
                <div className="d-flex align-items-center justify-content-between">
                    {value === null && nullLabel}
                    {value === undefined && placeholder}
                    {value && (
                        <Tooltip content={`${format(value[0], 'MMM d, yyyy')} - ${format(value[1], 'MMM d, yyyy')}`}>
                            <span>
                                {format(value[0], 'M/d')}-{format(value[1], 'M/d')}
                            </span>
                        </Tooltip>
                    )}
                    <Icon aria-label="Arrow down" svgPath={mdiChevronDown} className="ml-1" />
                </div>
            </PopoverTrigger>
            <PopoverContent>
                <Container className="d-flex flex-column">
                    <div className="d-flex justify-content-start mb-2">
                        {predefinedDates.map(([start, end, label]) => (
                            <Button
                                size="sm"
                                variant="secondary"
                                className="mr-2"
                                outline={true}
                                key={label}
                                onClick={() => handleChange([start, end])}
                            >
                                {label}
                            </Button>
                        ))}
                        {nullLabel && (
                            <Button
                                className="mr-2"
                                size="sm"
                                variant="secondary"
                                outline={true}
                                onClick={() => handleChange(null)}
                            >
                                {nullLabel}
                            </Button>
                        )}
                        <Button
                            className="ml-auto"
                            size="sm"
                            variant="link"
                            outline={true}
                            onClick={() => handleChange(undefined)}
                        >
                            Clear
                            <Icon aria-hidden={true} className="ml-1" svgPath={mdiClose} />
                        </Button>
                    </div>
                    <Calendar
                        highlightToday={true}
                        className="border-0 p-0 m-0"
                        isRange={true}
                        value={value}
                        onChange={handleChange}
                    />
                </Container>
            </PopoverContent>
        </Popover>
    )
}
