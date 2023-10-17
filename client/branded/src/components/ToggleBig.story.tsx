import { useState } from 'react'

import { action } from '@storybook/addon-actions'
import type { Meta, StoryFn } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { ToggleBig } from './ToggleBig'

const onToggle = action('onToggle')

const config: Meta = {
    title: 'branded/ToggleBig',
    decorators: [story => <BrandedStory>{() => <div className="container mt-3 pb-3">{story()}</div>}</BrandedStory>],
}

export default config

export const Interactive: StoryFn = () => {
    const [value, setValue] = useState(false)

    const onToggle = (value: boolean) => setValue(value)

    return (
        <div className="d-flex align-items-center">
            <ToggleBig value={value} onToggle={onToggle} title="Hello" className="mr-2" /> Value is {String(value)}
        </div>
    )
}

Interactive.parameters = {
    chromatic: {
        disable: true,
    },
}

export const On: StoryFn = () => <ToggleBig value={true} onToggle={onToggle} />
export const Off: StoryFn = () => <ToggleBig value={false} onToggle={onToggle} />
export const DisabledOn: StoryFn = () => <ToggleBig value={true} disabled={true} onToggle={onToggle} />

DisabledOn.storyName = 'Disabled & on'

export const DisabledOff: StoryFn = () => <ToggleBig value={false} disabled={true} onToggle={onToggle} />

DisabledOff.storyName = 'Disabled & off'
