import { Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { TeamMembersPage } from './TeamMembersPage'
import { testContext } from './test-utils'

const config: Meta = {
    title: 'web/teams/TeamMembersPage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}
export default config

export const Default: Story = function Default() {
    return <WebStory initialEntries={['/teams/team-1/members']}>{() => <TeamMembersPage {...testContext} />}</WebStory>
}
