import type { MockedResponse } from '@apollo/client/testing'
import type { Meta, StoryFn } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'

import { WebStory } from '../../components/WebStory'
import type { ListTeamsResult } from '../../graphql-operations'

import { LIST_TEAMS } from './backend'
import { TeamListPage } from './TeamListPage'

const config: Meta = {
    title: 'web/teams/TeamListPage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}
export default config

export const EmptyList: StoryFn = function EmptyList() {
    const mockResponse: MockedResponse<ListTeamsResult> = {
        request: {
            query: getDocumentNode(LIST_TEAMS),
        },
        result: {
            data: {
                __typename: 'Query',
                teams: {
                    __typename: 'TeamConnection',
                    nodes: [],
                    pageInfo: {
                        __typename: 'PageInfo',
                        endCursor: null,
                        hasNextPage: false,
                    },
                    totalCount: 0,
                },
            },
        },
    }

    return <WebStory mocks={[mockResponse]}>{() => <TeamListPage />}</WebStory>
}

export const ListWithItems: StoryFn = function ListWithItems() {
    const mockResponse: MockedResponse<ListTeamsResult> = {
        request: {
            query: getDocumentNode(LIST_TEAMS),
        },
        result: {
            data: {
                __typename: 'Query',
                teams: {
                    __typename: 'TeamConnection',
                    nodes: [
                        {
                            __typename: 'Team',
                            id: 'team-1',
                            name: 'team-1',
                            displayName: 'Team 1',
                            avatarURL: null,
                            url: '/teams/team-1',
                            readonly: false,
                            viewerCanAdminister: true,
                            parentTeam: null,
                            childTeams: {
                                __typename: 'TeamConnection',
                                totalCount: 1,
                            },
                            members: {
                                __typename: 'TeamMemberConnection',
                                totalCount: 1,
                            },
                        },
                        {
                            __typename: 'Team',
                            id: 'team-2',
                            name: 'team-2',
                            displayName: 'Team 2',
                            avatarURL: null,
                            url: '/teams/team-2',
                            readonly: true,
                            viewerCanAdminister: false,
                            parentTeam: null,
                            childTeams: {
                                __typename: 'TeamConnection',
                                totalCount: 0,
                            },
                            members: {
                                __typename: 'TeamMemberConnection',
                                totalCount: 5,
                            },
                        },
                    ],
                    pageInfo: {
                        __typename: 'PageInfo',
                        endCursor: null,
                        hasNextPage: false,
                    },
                    totalCount: 2,
                },
            },
        },
    }

    return <WebStory mocks={[mockResponse]}>{() => <TeamListPage />}</WebStory>
}
