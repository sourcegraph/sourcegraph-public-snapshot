import { action } from '@storybook/addon-actions'
import React, { useState } from 'react'

import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { ToggleBig } from './ToggleBig'

const onToggle = action('onToggle')

export default {
    title: 'branded/ToggleBig',

    decorators: [
        story => (
            <>
                <div>{story()}</div>
                <style>{webStyles}</style>
            </>
        ),
    ],
}

export const Interactive = () => {
    const [value, setValue] = useState(false)

    const onToggle = (value: boolean) => setValue(value)

    return (
        <div className="d-flex align-items-center">
            <ToggleBig value={value} onToggle={onToggle} title="Hello" className="mr-2" /> Value is {String(value)}
        </div>
    )
}

Interactive.story = {
    parameters: {
        chromatic: {
            disable: true,
        },
    },
}

export const On = () => <ToggleBig value={true} onToggle={onToggle} />
export const Off = () => <ToggleBig value={false} onToggle={onToggle} />
export const DisabledOn = () => <ToggleBig value={true} disabled={true} onToggle={onToggle} />

DisabledOn.story = {
    name: 'Disabled & on',
}

export const DisabledOff = () => <ToggleBig value={false} disabled={true} onToggle={onToggle} />

DisabledOff.story = {
    name: 'Disabled & off',
}
