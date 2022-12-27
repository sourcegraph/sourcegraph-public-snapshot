import { useState } from 'react'

import { action } from '@storybook/addon-actions'
import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { ToggleBig } from './ToggleBig'

const onToggle = action('onToggle')

const config: Meta = {
    title: 'branded/ToggleBig',
    decorators: [story => <BrandedStory>{() => <div className="container mt-3 pb-3">{story()}</div>}</BrandedStory>],
}

export default config

export const Interactive: Story = () => {
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

// eslint-disable-next-line id-length
export const On: Story = () => <ToggleBig value={true} onToggle={onToggle} />
export const Off: Story = () => <ToggleBig value={false} onToggle={onToggle} />
export const DisabledOn: Story = () => <ToggleBig value={true} disabled={true} onToggle={onToggle} />

DisabledOn.storyName = 'Disabled & on'

export const DisabledOff: Story = () => <ToggleBig value={false} disabled={true} onToggle={onToggle} />

DisabledOff.storyName = 'Disabled & off'
