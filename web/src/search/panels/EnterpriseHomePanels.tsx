import * as React from 'react'
import { AuthenticatedUser } from '../../auth'
import { RecentFilesPanel } from './RecentFilesPanel'
import { RecentSearchesPanel } from './RecentSearchesPanel'
import { RepositoriesPanel } from './RepositoriesPanel'
import { SavedSearchesPanel } from './SavedSearchesPanel'
import { Observable } from 'rxjs'
import { EventLogResult } from '../backend'
import { PatternTypeProps } from '..'
import * as GQL from '../../../../shared/src/graphql/schema'

interface Props extends Pick<PatternTypeProps, 'patternType'> {
    authenticatedUser: AuthenticatedUser | null
    fetchSavedSearches: () => Observable<GQL.ISavedSearch[]>
    fetchRecentSearches: (userId: string, first: number) => Observable<EventLogResult | null>
    fetchRecentFileViews: (userId: string, first: number) => Observable<EventLogResult | null>
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
