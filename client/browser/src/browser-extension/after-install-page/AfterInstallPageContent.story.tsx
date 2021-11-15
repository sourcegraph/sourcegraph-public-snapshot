import { Meta, Story } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'

import brandedStyles from '../../branded.scss'

import { AfterInstallPageContent } from './AfterInstallPageContent'

const config: Meta = {
    title: 'browser/AfterInstallPage',
}

export default config

export const Default: Story = () => <BrandedStory styles={brandedStyles}>{AfterInstallPageContent}</BrandedStory>
