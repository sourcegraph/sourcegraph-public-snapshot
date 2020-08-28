import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import React, { useState } from 'react'
import { Toggle } from './Toggle'
import webStyles from '../../../web/src/main.scss'

const onToggle = action('onToggle')

const { add } = storiesOf('shared/Toggle', module).addDecorator(story => (
    <>
        <div>{story()}</div>
        <style>{webStyles}</style>
    </>
))

add('Interactive', () => {
    const ToggleInteractive: React.FC = () => {
        const [value, setValue] = useState(false)

        const onToggle = (value: boolean) => setValue(value)

        return (
            <div className="d-flex align-items-center">
                <Toggle value={value} onToggle={onToggle} title="Hello" className="mr-2" /> Value is {String(value)}
            </div>
        )
    }

    return <ToggleInteractive />
})

add('On', () => <Toggle value={true} onToggle={onToggle} />)

add('Off', () => <Toggle value={false} onToggle={onToggle} />)

add('Disabled & on', () => <Toggle value={true} disabled={true} onToggle={onToggle} />)

add('Disabled & off', () => <Toggle value={false} disabled={true} onToggle={onToggle} />)
