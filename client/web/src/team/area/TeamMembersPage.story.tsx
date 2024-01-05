import type { MockedResponse } from '@apollo/client/testing'
import type { Meta, StoryFn } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'

import { WebStory } from '../../components/WebStory'
import type { ListTeamMembersResult } from '../../graphql-operations'
import { LIST_TEAM_MEMBERS } from '../members/backend'

import { TeamMembersPage } from './TeamMembersPage'
import { testContext } from './testContext.mock'

const config: Meta = {
    title: 'web/teams/TeamMembersPage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}
export default config

const mockResponse: MockedResponse<ListTeamMembersResult> = {
    request: {
        query: getDocumentNode(LIST_TEAM_MEMBERS),
    },
    result: {
        data: {
            __typename: 'Query',
            team: {
                __typename: 'Team',
                id: 'team-1',
                members: {
                    __typename: 'TeamMemberConnection',
                    totalCount: 2,
                    pageInfo: {
                        __typename: 'PageInfo',
                        endCursor: null,
                        hasNextPage: false,
                    },
                    nodes: [
                        {
                            __typename: 'User',
                            id: 'user-1',
                            username: 'user-1',
                            displayName: 'User 1',
                            avatarURL: null,
                            url: '/users/user-1',
                        },
                        {
                            __typename: 'User',
                            id: 'user-2',
                            username: 'user-2',
                            displayName: 'User 2',
                            avatarURL: null,
                            url: '/users/user-2',
                        },
                    ],
                },
            },
        },
    },
}

export const Default: StoryFn = function Default() {
    return (
        <WebStory initialEntries={['/teams/team-1/members']} mocks={[mockResponse]}>
            {() => <TeamMembersPage {...testContext} />}
        </WebStory>
    )
}
