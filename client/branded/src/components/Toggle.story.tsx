import { useState } from 'react'

import { action } from '@storybook/addon-actions'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Toggle } from './Toggle'

const ToggleExample: typeof Toggle = ({ value, disabled, onToggle }) => (
    <div className="d-flex align-items-baseline mb-2">
        <Toggle value={value} onToggle={onToggle} disabled={disabled} title="Hello" className="mr-2" />
        <div>
            <label className="mb-0">
                {disabled ? 'Disabled ' : ''}Toggle {value ? 'on' : 'off'}
            </label>
            <small className="field-message mt-0">This is helper text as needed</small>
        </div>
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
