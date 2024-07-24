import React, { useCallback, useState } from 'react'

import type { Meta } from '@storybook/react'

import { H1 } from '../..'
import { BrandedStory } from '../../../stories/BrandedStory'
import { Grid } from '../../Grid'

import { TextArea } from './TextArea'

const config: Meta = {
    title: 'wildcard/TextArea',

    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],

    parameters: {
        component: TextArea,

        design: {
            type: 'figma',
            name: 'Figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=860%3A79961',
        },
    },
}

export default config

export const TextAreaExamples: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => {
    const [value, setValue] = useState('')

    const handleChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(event => {
        setValue(event.target.value)
    }, [])

    return (
        <>
            <H1>TextArea</H1>
            <Grid columnCount={4}>
                <div>
                    <TextArea
                        onChange={handleChange}
                        value={value}
                        label="Standard example"
                        placeholder="Please type here..."
                    />
                </div>
                <div>
                    <TextArea
                        value=""
                        label="Disabled example"
                        disabled={true}
                        message="This is helper text as needed."
                        placeholder="Please type here..."
                    />
                </div>
                <div>
                    <TextArea
                        onChange={handleChange}
                        isValid={false}
                        value={value}
                        label="Error example"
                        message="show an error message"
                        placeholder="Please type here..."
                    />
                </div>
                <div>
                    <TextArea
                        onChange={handleChange}
                        value={value}
                        label="Small example"
                        placeholder="Please type here..."
                        size="small"
                    />
                </div>
            </Grid>
        </>
    )
}
