import { useState } from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'
import { addDays, startOfDay, subDays } from 'date-fns'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Calendar } from './Calendar'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
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
