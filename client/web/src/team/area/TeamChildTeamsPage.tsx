import { Page } from '../../components/Page'
import type { TeamAreaTeamFields } from '../../graphql-operations'
import { ChildTeamListPage } from '../list/TeamListPage'

import { TeamHeader } from './TeamHeader'

export interface TeamChildTeamsPageProps {
    /** The team that is the subject of the page. */
    team: TeamAreaTeamFields
}

export const TeamChildTeamsPage: React.FunctionComponent<TeamChildTeamsPageProps> = ({ team }) => (
    <Page>
        <TeamHeader team={team} className="mb-3" />
        <div className="container">
            <ChildTeamListPage parentTeam={team.name} />
        </div>
    </Page>
)
