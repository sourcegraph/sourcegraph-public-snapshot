import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'

import { AfterInstallPageContent } from './AfterInstallPageContent'

import brandedStyles from '../../branded.scss'

const config: Meta = {
    title: 'browser/AfterInstallPage',
    parameters: {
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default config

export const Default: Story = () => <BrandedStory styles={brandedStyles}>{AfterInstallPageContent}</BrandedStory>
