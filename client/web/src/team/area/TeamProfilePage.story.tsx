import type { Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { TeamProfilePage } from './TeamProfilePage'
import { testContext } from './testContext.mock'

const config: Meta = {
    title: 'web/teams/TeamProfilePage',
    parameters: {},
}
export default config

export const Default: StoryFn = function Default() {
    return <WebStory initialEntries={['/teams/team-1']}>{() => <TeamProfilePage {...testContext} />}</WebStory>
}
