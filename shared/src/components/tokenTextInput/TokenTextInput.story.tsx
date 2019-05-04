import { action } from '@storybook/addon-actions'
import { text } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import React, { useState } from 'react'
import { TokenTextInput } from './TokenTextInput'

const onChange = action('onChange')

// Disable keyboard shortcuts because in TokenTextInput the cursor is in a contenteditable element,
// which Storybook doesn't consider to be an input, so it intercepts keyboard shortcuts instead of
// propagating them to the TokenTextInput element.
const { add } = storiesOf('TokenTextInput', module).addParameters({ options: { enableShortcuts: false } })

// tslint:disable: jsx-no-lambda

add('interactive', () => {
    const TokenTextInputInteractive: React.FunctionComponent = () => {
        const [value, setValue] = useState<string>(text('Initial value', 'mango tomato onio'))
        return (
            <TokenTextInput
                className="bg-white mt-5 m-4 p-1"
                value={value}
                onChange={newValue => {
                    onChange(newValue)
                    setValue(newValue)
                }}
            />
        )
    }
    return <TokenTextInputInteractive />
})
