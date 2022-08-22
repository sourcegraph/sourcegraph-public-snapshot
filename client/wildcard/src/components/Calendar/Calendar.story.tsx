import React from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'
import { startOfYesterday } from 'date-fns'

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

export const Default: Story = () => {
    const [value, onChange] = React.useState(new Date())
    return <Calendar value={value} onChange={onChange} />
}

Default.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
    design: {
        type: 'figma',
        name: 'Figma',
        url: '#',
    },
}

export const ModeRange: Story = () => {
    const [value, onChange] = React.useState([startOfYesterday(), new Date()])
    return <Calendar mode="range" value={value as [Date, Date]} onChange={onChange} />
}

ModeRange.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
    design: {
        type: 'figma',
        name: 'Figma',
        url: '#',
    },
}
