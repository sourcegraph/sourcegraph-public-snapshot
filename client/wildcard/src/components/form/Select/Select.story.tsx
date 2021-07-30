import { Meta } from '@storybook/react'
import React, { useCallback } from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Select, SelectProps } from './Select'

const Story: Meta = {
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
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A1353',
        },
    },
}

// eslint-disable-next-line import/no-default-export
export default Story

const BaseSelect = (props: Partial<SelectProps>) => {
    const [selected, setSelected] = React.useState('')

    const handleChange = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(event => {
        setSelected(event.target.value)
    }, [])

    return (
        <Select label="What is your favorite fruit?" value={selected} onChange={handleChange} {...props}>
            <option value="">Favorite fruit</option>
            <option value="apples">Apples</option>
            <option value="bananas">Bananas</option>
            <option value="oranges">Oranges</option>
        </Select>
    )
}

const SelectVariants = ({ isCustomStyle }: Pick<SelectProps, 'isCustomStyle'>) => (
    <>
        <h3>Simple</h3>
        <BaseSelect isCustomStyle={isCustomStyle} />
        <BaseSelect isCustomStyle={isCustomStyle} isValid={false} />
        <BaseSelect isCustomStyle={isCustomStyle} isValid={true} />

        <h3>With message</h3>
        <BaseSelect isCustomStyle={isCustomStyle} message="I am a message" />
        <BaseSelect isCustomStyle={isCustomStyle} message="I am a message" isValid={false} />
        <BaseSelect isCustomStyle={isCustomStyle} message="I am a message" isValid={true} />
    </>
)

export const SelectExamples: React.FunctionComponent = () => (
    <>
        <h1>Select</h1>
        <h2>Native select</h2>
        <SelectVariants />
        <h2>Custom select</h2>
        <SelectVariants isCustomStyle={true} />
    </>
)
