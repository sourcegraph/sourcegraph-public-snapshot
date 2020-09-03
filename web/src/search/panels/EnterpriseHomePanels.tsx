import * as React from 'react'
import { AuthenticatedUser } from '../../auth'
import { RecentFilesPanel } from './RecentFilesPanel'
import { RecentSearchesPanel } from './RecentSearchesPanel'
import { RepositoriesPanel } from './RepositoriesPanel'
import { SavedSearchesPanel } from './SavedSearchesPanel'

import { PatternTypeProps } from '..'
import { Observable } from 'rxjs'
import * as GQL from '../../../../shared/src/graphql/schema'

interface Props extends Pick<PatternTypeProps, 'patternType'> {
    authenticatedUser: AuthenticatedUser | null
    fetchSavedSearches: () => Observable<GQL.ISavedSearch[]>
}

export const EnterpriseHomePanels: React.FunctionComponent<Props> = (props: Props) => (
    <div className="enterprise-home-panels container">
        <div className="row">
            <RepositoriesPanel className="enterprise-home-panels__panel col-lg-4" />
            <RecentSearchesPanel
                authenticatedUser={authenticatedUser}
                className="enterprise-home-panels__panel col-lg-8"
            />
        </div>
        <div className="row">
            <RecentFilesPanel className="enterprise-home-panels__panel col-md-7" />
            <SavedSearchesPanel {...props} className="enterprise-home-panels__panel col-md-5" />
        </div>
    </div>
)
