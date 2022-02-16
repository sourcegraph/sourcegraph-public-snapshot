import { boolean } from '@storybook/addon-knobs'
import { Meta } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { LoadingSpinner } from './LoadingSpinner'

const config: Meta = {
    title: 'wildcard/LoadingSpinner',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: LoadingSpinner,
        chromatic: {
            enableDarkMode: true,
        },
    },
}

export default config

export const LoadingSpinnerExample = () => <LoadingSpinner inline={boolean('inline', true)} />
