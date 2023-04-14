import { Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { NewTeamPage } from './NewTeamPage'

const config: Meta = {
    title: 'web/teams/NewTeamPage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}
export default config

export const Default: Story = function Default() {
    return <WebStory>{() => <NewTeamPage />}</WebStory>
}
