import { useState } from 'react'

import { action } from '@storybook/addon-actions'
import type { Meta, StoryFn } from '@storybook/react'

import { Label } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { Toggle } from './Toggle'

const ToggleExample: typeof Toggle = ({ value, disabled, onToggle }) => (
    <div className="d-flex align-items-baseline mb-2">
        <Toggle value={value} onToggle={onToggle} disabled={disabled} title="Hello" className="mr-2" />
        <div>
            <Label className="mb-0">
                {disabled ? 'Disabled ' : ''}Toggle {value ? 'on' : 'off'}
            </Label>
            <small className="field-message mt-0">This is helper text as needed</small>
        </div>
    </div>
)
const onToggle = action('onToggle')

const config: Meta = {
    title: 'branded/Toggle',
    decorators: [story => <BrandedStory>{() => <div className="container mt-3 pb-3">{story()}</div>}</BrandedStory>],
}

export default config

export const Interactive: StoryFn = () => {
    const [value, setValue] = useState(false)

    const onToggle = (value: boolean) => setValue(value)

    return <ToggleExample value={value} onToggle={onToggle} />
}

export const Variants: StoryFn = () => (
    <>
        <ToggleExample value={true} onToggle={onToggle} />
        <ToggleExample value={false} onToggle={onToggle} />
        <ToggleExample value={true} disabled={true} onToggle={onToggle} />
        <ToggleExample value={false} disabled={true} onToggle={onToggle} />
    </>
)
