import { useState } from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'
import { addDays, startOfDay, subDays } from 'date-fns'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Badge, Text } from '..'

import { Calendar } from './Calendar'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>
        {() => (
            <div className="container mt-3">
                <Text>
                    This is an{' '}
                    <Badge variant="merged" className="mb-2">
                        Experimental
                    </Badge>{' '}
                    component and built on top of `react-calendar` package with Sourcegraph CSS styling on top. It
                    intentionally, omits other `react-calendar` props/features to not over-complicate and use as simple
                    calendar, in case if we migrate to another calendar library or build our own.
                </Text>
                <div className="d-flex flex-column">{story()}</div>
            </div>
        )}
    </BrandedStory>
)

const config: Meta = {
    title: 'wildcard/Calendar',
    component: Calendar,
    decorators: [decorator],
}

export default config

// NOTE: hardcoded in order to screenshot test the calendar
const today = startOfDay(new Date('2022-08-22'))

export const Single: Story = () => {
    const [value, onChange] = useState(today)
    return <Calendar value={value} onChange={onChange} />
}

Single.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}

export const Range: Story = () => {
    const [value, onChange] = useState<[Date, Date]>([subDays(today, 7), today])
    return <Calendar isRange={true} value={value} onChange={onChange} />
}

Range.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}

export const MinMaxDates: Story = () => {
    const [value, onChange] = useState<[Date, Date]>([subDays(today, 7), today])
    return (
        <Calendar
            isRange={true}
            value={value}
            onChange={onChange}
            minDate={subDays(today, 10)}
            maxDate={addDays(today, 10)}
        />
    )
}

MinMaxDates.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}

export const HighlightToday: Story = () => {
    const [value, onChange] = useState(new Date())
    return <Calendar value={value} onChange={onChange} highlightToday={true} />
}

HighlightToday.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: true,
    },
}

export const HighlightTodayRange: Story = () => {
    const [value, onChange] = useState<[Date, Date]>([subDays(new Date(), 4), addDays(new Date(), 3)])
    return <Calendar isRange={true} value={value} onChange={onChange} highlightToday={true} />
}

HighlightTodayRange.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: true,
    },
}
