import { Meta, Story } from '@storybook/react'
import React, { useCallback } from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Grid } from '../../Grid/Grid'

import { Select, SelectProps } from './Select'

const config: Meta = {
    title: 'wildcard/Select',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: Select,
        design: {
            type: 'figma',
            name: 'Figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=854%3A1630',
        },
    },
}

export default config

const BaseSelect = (props: { id: string } & Pick<SelectProps, 'isCustomStyle' | 'isValid' | 'disabled'>) => {
    const [selected, setSelected] = React.useState('')

    const handleChange = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(event => {
        setSelected(event.target.value)
    }, [])

    return (
        <Select
            label="What is your favorite fruit?"
            message="I am a message"
            value={selected}
            onChange={handleChange}
            {...props}
        >
            <option value="">Favorite fruit</option>
            <option value="apples">Apples</option>
            <option value="bananas">Bananas</option>
            <option value="oranges">Oranges</option>
        </Select>
    )
}

const SelectVariants = ({ isCustomStyle }: Pick<SelectProps, 'isCustomStyle'>) => {
    const idPrefix = isCustomStyle ? 'custom' : 'native'
    return (
        <Grid columnCount={4}>
            <div>
                <h2>Standard</h2>
                <BaseSelect id={`${idPrefix}-standard`} isCustomStyle={isCustomStyle} />
            </div>
            <div>
                <h2>Valid</h2>
                <BaseSelect id={`${idPrefix}-valid`} isCustomStyle={isCustomStyle} isValid={true} />
            </div>
            <div>
                <h2>Invalid</h2>
                <BaseSelect id={`${idPrefix}-invalid`} isCustomStyle={isCustomStyle} isValid={false} />
            </div>
            <div>
                <h2>Disabled</h2>
                <BaseSelect id={`${idPrefix}-disabled`} isCustomStyle={isCustomStyle} disabled={true} />
            </div>
        </Grid>
    )
}

export const SelectExamples: Story = () => (
    <>
        <h1>Select</h1>
        <h2>Native</h2>
        <SelectVariants />
        <h2>Custom</h2>
        <SelectVariants isCustomStyle={true} />
    </>
)

SelectExamples.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}
