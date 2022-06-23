import { Meta, Story } from '@storybook/react'

import { WebStory } from '../components/WebStory'

import { TosConsentModal } from './TosConsentModal'

const config: Meta = {
    title: 'web/auth/TosConsentModal',
}

export default config

export const Standard: Story = () => <WebStory>{() => <TosConsentModal afterTosAccepted={() => {}} />}</WebStory>
