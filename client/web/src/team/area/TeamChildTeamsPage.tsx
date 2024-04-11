import { useEffect } from 'react'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'

import { Page } from '../../components/Page'
import type { TeamAreaTeamFields } from '../../graphql-operations'
import { ChildTeamListPage } from '../list/TeamListPage'

import { TeamHeader } from './TeamHeader'

export interface TeamChildTeamsPageProps extends TelemetryV2Props {
    /** The team that is the subject of the page. */
    team: TeamAreaTeamFields
}

export const TeamChildTeamsPage: React.FunctionComponent<TeamChildTeamsPageProps> = ({ team, telemetryRecorder }) => {
    useEffect(() => telemetryRecorder.recordEvent('team.childTeams', 'view'), [telemetryRecorder])
    return (
        <Page>
            <TeamHeader team={team} className="mb-3" />
            <div className="container">
                <ChildTeamListPage parentTeam={team.name} telemetryRecorder={telemetryRecorder} />
            </div>
        </Page>
    )
}
