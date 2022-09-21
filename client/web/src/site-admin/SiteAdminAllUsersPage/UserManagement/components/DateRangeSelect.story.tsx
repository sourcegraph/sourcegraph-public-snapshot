import { DecoratorFn, Meta, Story } from '@storybook/react'
import { startOfDay, subDays } from 'date-fns'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { DateRangeSelect } from './DateRangeSelect'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'web/UserManagementPage/DateRangeSelect',
    component: DateRangeSelect,
    decorators: [decorator],
}

export default config

// NOTE: hardcoded in order to screenshot test the calendar
const today = startOfDay(new Date('2022-08-30'))
const defaultValue: [Date, Date] = [subDays(today, 7), today]

export const Default: Story = () => (
    <div className="d-flex justify-content-around w-50">
        <DateRangeSelect value={defaultValue} defaultIsOpen={true} />
    </div>
)

Default.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: true,
    },
}

export const WithNegation: Story = () => (
    <div className="d-flex justify-content-around w-50">
        <DateRangeSelect
            negation={{
                label: 'With negation',
                value: true,
                message: 'With negation description message',
            }}
            value={defaultValue}
            defaultIsOpen={true}
        />
    </div>
)

WithNegation.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: true,
    },
}
