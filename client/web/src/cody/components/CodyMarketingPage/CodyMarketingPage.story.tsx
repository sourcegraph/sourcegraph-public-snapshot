import { Meta, Story } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { CodyMarketingPage } from './CodyMarketingPage'

const config: Meta = {
    title: 'web/src/cody/CodyMarketingPage',
}

export default config

export const Default: Story = () => <WebStory>{() => <CodyMarketingPage />}</WebStory>
