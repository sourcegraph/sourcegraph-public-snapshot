import React, { useCallback } from 'react'

import type { Meta, StoryFn } from '@storybook/react'

import { H1, H2 } from '../..'
import { BrandedStory } from '../../../stories/BrandedStory'
import { Grid } from '../../Grid'

import { Checkbox, type CheckboxProps } from './Checkbox'

const config: Meta = {
    title: 'wildcard/Checkbox',

    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],

    parameters: {
        component: Checkbox,
        design: {
            type: 'figma',
            name: 'Figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=860%3A79469',
        },
    },
}

export default config

const BaseCheckbox = ({ name, ...props }: { name: string } & Pick<CheckboxProps, 'isValid' | 'disabled'>) => {
    const [isChecked, setChecked] = React.useState(false)

    const handleChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setChecked(event.target.checked)
    }, [])

    return (
        <Checkbox
            name={name}
            id={name}
            value="first"
            checked={isChecked}
            onChange={handleChange}
            label="Check me!"
            message="Hello world!"
            {...props}
        />
    )
}

export const CheckboxExamples: StoryFn = () => (
    <>
        <H1>Checkbox</H1>
        <Grid columnCount={4}>
            <div>
                <H2>Standard</H2>
                <BaseCheckbox name="standard-example" />
            </div>
            <div>
                <H2>Valid</H2>
                <BaseCheckbox name="valid-example" isValid={true} />
            </div>
            <div>
                <H2>Invalid</H2>
                <BaseCheckbox name="invalid-example" isValid={false} />
            </div>
            <div>
                <H2>Disabled</H2>
                <BaseCheckbox name="disabled-example" disabled={true} />
            </div>
        </Grid>
    </>
)
