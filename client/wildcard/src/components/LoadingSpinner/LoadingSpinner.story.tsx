import { boolean } from '@storybook/addon-knobs'
import { Meta } from '@storybook/react'
import React from 'react'

import { LoadingSpinner } from './LoadingSpinner'

const config: Meta = {
    title: 'wildcard/LoadingSpinner',

    parameters: {
        component: LoadingSpinner,
    },
}

// eslint-disable-next-line import/no-default-export
export default config

export const LoadingSpinnerExample = () => <LoadingSpinner inline={boolean('inline', true)} />
