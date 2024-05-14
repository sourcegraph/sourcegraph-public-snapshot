import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'

import { Page } from '../../components/Page'
import type { TeamAreaTeamFields } from '../../graphql-operations'
import { TeamMemberListPage } from '../members/TeamMemberListPage'

import { TeamHeader } from './TeamHeader'

export interface TeamMembersPageProps extends TelemetryV2Props {
    /** The team that is the subject of the page. */
    team: TeamAreaTeamFields
}

export const TeamMembersPage: React.FunctionComponent<TeamMembersPageProps> = ({ team, telemetryRecorder }) => (
    <Page>
        <TeamHeader team={team} className="mb-3" />
        <div className="container">
            <TeamMemberListPage
                teamID={team.id}
                teamName={team.name}
                viewerCanAdminister={team.viewerCanAdminister}
                telemetryRecorder={telemetryRecorder}
            />
        </div>
    </Page>
)
