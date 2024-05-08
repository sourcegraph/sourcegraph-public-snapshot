import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { mockAuthenticatedUser } from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'

import type { TeamAreaRouteContext } from './TeamArea'

export const testContext: TeamAreaRouteContext = {
    authenticatedUser: mockAuthenticatedUser,
    team: {
        __typename: 'Team',
        id: '1',
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
            totalCount: 2,
        },
        creator: {
            __typename: 'User',
            username: 'alice',
            displayName: 'Alice',
            avatarURL: null,
            url: '/users/alice',
        },
    },
    telemetryRecorder: noOpTelemetryRecorder,
    onTeamUpdate: () => {},
}
