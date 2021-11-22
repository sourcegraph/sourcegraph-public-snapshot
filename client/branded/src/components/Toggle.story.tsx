import { action } from '@storybook/addon-actions'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import React, { useState } from 'react'

import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Toggle } from './Toggle'

const ToggleExample: typeof Toggle = ({ value, disabled, onToggle }) => (
    <div className="mb-2">
        <Toggle
            value={value}
            onToggle={onToggle}
            disabled={disabled}
            title="Hello"
            label={(disabled ? "Disabled" : "") + " Toggle on"}
            offLabel={(disabled ? "Disabled" : "") + " Toggle off"}
            helpText="This is helper text as needed" />
    </div>
)
const onToggle = action('onToggle')

const decorator: DecoratorFn = story => (
    <>
        <div>{story()}</div>
        <style>{webStyles}</style>
    </>
)
const config: Meta = {
    title: 'branded/Toggle',
    decorators: [decorator],
}

export default config

export const Interactive: Story = () => {
    const [value, setValue] = useState(false)

    const onToggle = (value: boolean) => setValue(value)

    return <ToggleExample value={value} onToggle={onToggle} />
}

Interactive.parameters = {
    chromatic: {
        disable: true,
    },
}

export const Variants: Story = () => (
    <>
        <ToggleExample value={true} onToggle={onToggle} />
        <ToggleExample value={false} onToggle={onToggle} />
        <ToggleExample value={true} disabled={true} onToggle={onToggle} />
        <ToggleExample value={false} disabled={true} onToggle={onToggle} />
    </>
)
