import * as React from 'react'
import { AuthenticatedUser } from '../../auth'
import { EnterpriseHomePanelsProps, PatternTypeProps } from '..'
import { RecentFilesPanel } from './RecentFilesPanel'
import { RecentSearchesPanel } from './RecentSearchesPanel'
import { RepositoriesPanel } from './RepositoriesPanel'
import { SavedSearchesPanel } from './SavedSearchesPanel'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'

interface Props extends Pick<PatternTypeProps, 'patternType'>, TelemetryProps, EnterpriseHomePanelsProps {
    authenticatedUser: AuthenticatedUser | null
}

export const EnterpriseHomePanels: React.FunctionComponent<Props> = (props: Props) => (
    <div className="enterprise-home-panels container">
        <div className="row">
            <RepositoriesPanel {...props} className="enterprise-home-panels__panel col-lg-4" />
            <RecentSearchesPanel {...props} className="enterprise-home-panels__panel col-lg-8" />
        </div>
        <div className="row">
            <RecentFilesPanel {...props} className="enterprise-home-panels__panel col-lg-7" />
            <SavedSearchesPanel {...props} className="enterprise-home-panels__panel col-lg-5" />
        </div>
    </div>
)
