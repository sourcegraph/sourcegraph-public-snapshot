import { Page } from '../../components/Page'
import type { TeamAreaTeamFields } from '../../graphql-operations'
import { TeamMemberListPage } from '../members/TeamMemberListPage'

import { TeamHeader } from './TeamHeader'

export interface TeamMembersPageProps {
    /** The team that is the subject of the page. */
    team: TeamAreaTeamFields
}

export const TeamMembersPage: React.FunctionComponent<TeamMembersPageProps> = ({ team }) => (
    <Page>
        <TeamHeader team={team} className="mb-3" />
        <div className="container">
            <TeamMemberListPage teamID={team.id} teamName={team.name} viewerCanAdminister={team.viewerCanAdminister} />
        </div>
    </Page>
)
