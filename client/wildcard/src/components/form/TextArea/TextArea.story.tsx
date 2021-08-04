import { Meta } from '@storybook/react'
import { StoryFnReactReturnType } from '@storybook/react/dist/ts3.9/client/preview/types'
import React, { useCallback, useState } from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Grid } from '../../Grid'

import { TextArea } from './TextArea'

const config: Meta = {
    title: 'wildcard/TextArea',

    decorators: [
        (story: () => StoryFnReactReturnType): StoryFnReactReturnType => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: TextArea,
        design: {
            type: 'figma',
            name: 'Figma',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A1943',
        },
    },
}

// eslint-disable-next-line import/no-default-export
export default config

export const TextAreaExamples: React.FunctionComponent = () => {
    const [value, setValue] = useState('')

    const handleChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(event => {
        setValue(event.target.value)
    }, [])

    return (
        <>
            <h1>TextArea</h1>
            <Grid columnCount={3}>
                <div>
                    <TextArea
                        onChange={handleChange}
                        value={value}
                        title="Standard example"
                        placeholder="Please type here..."
                    />
                </div>
                <div>
                    <TextArea
                        value=""
                        title="Disabled example"
                        disabled={true}
                        message="This is helper text as needed."
                        placeholder="Please type here..."
                    />
                </div>
                <div>
                    <TextArea
                        onChange={handleChange}
                        isError={true}
                        value={value}
                        title="Error example"
                        message="show an error message"
                        placeholder="Please type here..."
                    />
                </div>
            </Grid>
        </>
    )
}
