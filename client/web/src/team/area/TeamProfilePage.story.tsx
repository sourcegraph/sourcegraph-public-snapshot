import { Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { TeamProfilePage } from './TeamProfilePage'
import { testContext } from './test-utils'

const config: Meta = {
    title: 'web/teams/TeamProfilePage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}
export default config

export const Default: Story = function Default() {
    return <WebStory initialEntries={['/teams/team-1']}>{() => <TeamProfilePage {...testContext} />}</WebStory>
}
