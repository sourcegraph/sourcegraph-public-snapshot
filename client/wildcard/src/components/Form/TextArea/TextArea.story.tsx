import React, { useCallback, useState } from 'react'

import { Meta } from '@storybook/react'
import { StoryFnReactReturnType } from '@storybook/react/dist/ts3.9/client/preview/types'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Typography } from '../..'
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
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
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
            <Typography.H1>TextArea</Typography.H1>
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
