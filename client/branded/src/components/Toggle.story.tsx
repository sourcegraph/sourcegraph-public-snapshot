import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import React, { useState } from 'react'

import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Toggle } from './Toggle'

const onToggle = action('onToggle')

const { add } = storiesOf('branded/Toggle', module).addDecorator(story => (
    <>
        <div>{story()}</div>
        <style>{webStyles}</style>
    </>
))

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

add(
    'Interactive',
    () => {
        const [value, setValue] = useState(false)

        const onToggle = (value: boolean) => setValue(value)

        return <ToggleExample value={value} onToggle={onToggle} />
    },
    {
        chromatic: {
            disable: true,
        },
    }
)

add('Variants', () => (
    <>
        <ToggleExample value={true} onToggle={onToggle} />
        <ToggleExample value={false} onToggle={onToggle} />
        <ToggleExample value={true} disabled={true} onToggle={onToggle} />
        <ToggleExample value={false} disabled={true} onToggle={onToggle} />
    </>
))
