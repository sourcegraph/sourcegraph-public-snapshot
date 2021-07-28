import { Meta } from '@storybook/react'
import React, { useCallback } from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { RadioButton } from './RadioButton'

const Story: Meta = {
    title: 'wildcard/RadioButton',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: RadioButton,
        design: {
            type: 'figma',
            name: 'Figma',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A1943',
        },
    },
}

// eslint-disable-next-line import/no-default-export
export default Story

export const Simple = () => {
    const [selected, setSelected] = React.useState('')

    const handleChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setSelected(event.target.value)
    }, [])

    return (
        <>
            <RadioButton
                value="first"
                checked={selected === 'first'}
                onChange={handleChange}
                label="First"
                validationMessage="Hello world!"
            />
            <RadioButton
                value="second"
                checked={selected === 'second'}
                onChange={handleChange}
                label="Second"
                validationMessage="Hello world!"
            />
            <RadioButton
                value="third"
                checked={selected === 'third'}
                onChange={handleChange}
                label="Third"
                validationMessage="Hello world!"
            />
        </>
    )
}
