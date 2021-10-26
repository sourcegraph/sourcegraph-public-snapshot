import { boolean } from '@storybook/addon-knobs'
import React from 'react'

import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { BrandedStory } from './BrandedStory'
import { LoaderInput } from './LoaderInput'

export default {
    title: 'branded/LoaderInput',

    decorators: [
        story => (
            <div className="container mt-3" style={{ width: 800 }}>
                {story()}
            </div>
        ),
    ],
}

export const Interactive = () => (
    <BrandedStory styles={webStyles}>
        {() => (
            <LoaderInput loading={boolean('loading', true)}>
                <input type="text" placeholder="Loader input" className="form-control" />
            </LoaderInput>
        )}
    </BrandedStory>
)
