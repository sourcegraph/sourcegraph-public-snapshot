import { Meta } from '@storybook/react'
import { StoryFnReactReturnType } from '@storybook/react/dist/ts3.9/client/preview/types'
import React, { useCallback } from 'react'

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

export const TextAreaExamples: React.FunctionComponent = () => (
    <>
        <h1>Radio</h1>
        <Grid columnCount={4}>
            <div>
                <h2>Standard</h2>
                <TextArea value="" title="Standard example" />
                <TextArea status="error" value="error example" title="Error example" />
            </div>
        </Grid>
    </>
)
