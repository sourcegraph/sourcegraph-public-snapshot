import React, { useCallback } from 'react'

import type { Meta, StoryFn } from '@storybook/react'

import { H1, H2 } from '../..'
import { BrandedStory } from '../../../stories/BrandedStory'
import { Grid } from '../../Grid'

import { RadioButton, type RadioButtonProps } from './RadioButton'

const config: Meta = {
    title: 'wildcard/RadioButton',

    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],

    parameters: {
        component: RadioButton,
        design: {
            type: 'figma',
            name: 'Figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=854%3A2792',
        },
    },
}

export default config

const BaseRadio = ({ name, ...props }: Pick<RadioButtonProps, 'name' | 'isValid' | 'disabled'>) => {
    const [selected, setSelected] = React.useState('')

    const handleChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setSelected(event.target.value)
    }, [])

    return (
        <>
            <RadioButton
                id={`${name}-1`}
                name={name}
                value="first"
                checked={selected === 'first'}
                onChange={handleChange}
                label="First"
                message="Hello world!"
                {...props}
            />
            <RadioButton
                id={`${name}-2`}
                name={name}
                value="second"
                checked={selected === 'second'}
                onChange={handleChange}
                label="Second"
                message="Hello world!"
                {...props}
            />
            <RadioButton
                id={`${name}-3`}
                name={name}
                value="third"
                checked={selected === 'third'}
                onChange={handleChange}
                label="Third"
                message="Hello world!"
                {...props}
            />
        </>
    )
}

export const RadioExamples: StoryFn = () => (
    <>
        <H1>Radio</H1>
        <Grid columnCount={4}>
            <div>
                <H2>Standard</H2>
                <BaseRadio name="standard-example" />
            </div>
            <div>
                <H2>Valid</H2>
                <BaseRadio name="valid-example" isValid={true} />
            </div>
            <div>
                <H2>Invalid</H2>
                <BaseRadio name="invalid-example" isValid={false} />
            </div>
            <div>
                <H2>Disabled</H2>
                <BaseRadio name="disabled-example" disabled={true} />
            </div>
        </Grid>
    </>
)

RadioExamples.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}
