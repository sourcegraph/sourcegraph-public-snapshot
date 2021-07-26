import { boolean, select } from '@storybook/addon-knobs'
import { Meta } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { BUTTON_VARIANTS, BUTTON_SIZES } from '../Button/constants'

import { LoadingButton } from './LoadingButton'

const Story: Meta = {
    title: 'wildcard/Buttons/LoadingButton',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: LoadingButton,
    },
}

// eslint-disable-next-line import/no-default-export
export default Story

export const Simple = () => (
    <LoadingButton
        loading={boolean('Loading', true)}
        alwaysShowChildren={boolean('Always show label', true)}
        variant={select('Variant', BUTTON_VARIANTS, 'secondary')}
        size={select('Size', BUTTON_SIZES, undefined)}
        disabled={boolean('Disabled', false)}
        outline={boolean('Outline', false)}
    >
        Click me!
    </LoadingButton>
)
