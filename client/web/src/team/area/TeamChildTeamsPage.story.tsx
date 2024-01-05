import type { MockedResponse } from '@apollo/client/testing'
import type { Meta, StoryFn } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'

import { WebStory } from '../../components/WebStory'
import type { ListTeamsOfParentResult } from '../../graphql-operations'
import { LIST_TEAMS_OF_PARENT } from '../list/backend'

import { TeamChildTeamsPage } from './TeamChildTeamsPage'
import { testContext } from './testContext.mock'

const config: Meta = {
    title: 'web/teams/TeamChildTeamsPage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}
export default config

const mockResponse: MockedResponse<ListTeamsOfParentResult> = {
    request: {
        query: getDocumentNode(LIST_TEAMS_OF_PARENT),
    },
    result: {
        data: {
            __typename: 'Query',
            team: {
                __typename: 'Team',
                childTeams: {
                    __typename: 'TeamConnection',
                    totalCount: 2,
                    pageInfo: {
                        __typename: 'PageInfo',
                        endCursor: null,
                        hasNextPage: false,
                    },
                    nodes: [
                        {
                            __typename: 'Team',
                            id: 'team-2',
                            name: 'team-2',
                            displayName: 'Team 2',
                            url: '/teams/team-2',
                            avatarURL: null,
                            readonly: false,
                            viewerCanAdminister: true,
                            parentTeam: null,
                            childTeams: {
                                __typename: 'TeamConnection',
                                totalCount: 0,
                            },
                            members: {
                                __typename: 'TeamMemberConnection',
                                totalCount: 2,
                            },
                        },
                        {
                            __typename: 'Team',
                            id: 'team-3',
                            name: 'team-3',
                            displayName: 'Team 3',
                            url: '/teams/team-3',
                            avatarURL: null,
                            readonly: false,
                            viewerCanAdminister: true,
                            parentTeam: null,
                            childTeams: {
                                __typename: 'TeamConnection',
                                totalCount: 1,
                            },
                            members: {
                                __typename: 'TeamMemberConnection',
                                totalCount: 0,
                            },
                        },
                    ],
                },
            },
        },
    },
}

export const Default: StoryFn = function Default() {
    return (
        <WebStory mocks={[mockResponse]} initialEntries={['/teams/team-1/child-teams']}>
            {() => <TeamChildTeamsPage {...testContext} />}
        </WebStory>
    )
}
