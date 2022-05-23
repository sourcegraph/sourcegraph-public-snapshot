import React, { useCallback } from 'react'

import { Meta } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { FormInput } from './FormInput'

const Story: Meta = {
    title: 'wildcard/FormInput',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: FormInput,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
        design: {
            type: 'figma',
            name: 'Figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=875%3A797',
        },
    },
}

export default Story

export const Simple = () => {
    const [selected, setSelected] = React.useState('')

    const handleChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setSelected(event.target.value)
    }, [])

    return (
        <>
            <FormInput
                value={selected}
                label="Input loading"
                onChange={handleChange}
                message="random message"
                status="loading"
                disabled={false}
                placeholder="testing this one"
            />

            <FormInput label="Input raw" value={selected} onChange={handleChange} />

            <FormInput
                value={selected}
                label="Input valid"
                onChange={handleChange}
                message="random message"
                status="valid"
                disabled={false}
                placeholder="testing this one"
            />

            <FormInput
                value={selected}
                label="Input error"
                onChange={handleChange}
                error="a message with error"
                status="error"
                placeholder="error status input"
            />
            <FormInput
                value={selected}
                label="Disabled input"
                onChange={handleChange}
                message="random message"
                disabled={true}
                placeholder="disable status input"
            />

            <FormInput
                value={selected}
                label="Input small"
                onChange={handleChange}
                message="random message"
                status="valid"
                disabled={false}
                placeholder="testing this one"
                variant="small"
            />
        </>
    )
}
