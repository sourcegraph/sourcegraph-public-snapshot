import { Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { TeamChildTeamsPage } from './TeamChildTeamsPage'
import { testContext } from './test-utils'

const config: Meta = {
    title: 'web/teams/TeamChildTeamsPage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}
export default config

export const Default: Story = function Default() {
    return (
        <WebStory initialEntries={['/teams/team-1/child-teams']}>
            {() => <TeamChildTeamsPage {...testContext} />}
        </WebStory>
    )
}
