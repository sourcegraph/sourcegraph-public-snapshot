import * as React from 'react'
import { RecentFilesPanel } from './RecentFilesPanel'
import { RecentSearchesPanel } from './RecentSearchesPanel'
import { RepositoriesPanel } from './RepositoriesPanel'
import { SavedSearchesPanel } from './SavedSearchesPanel'

export const EnterpriseHomePanels: React.FunctionComponent<{}> = () => (
    <div className="enterprise-home-panels container">
        <div className="row">
            <RepositoriesPanel className="enterprise-home-panels__panel col-lg-4" />
            <RecentSearchesPanel className="enterprise-home-panels__panel col-lg-8" />
        </div>
        <div className="row">
            <RecentFilesPanel className="enterprise-home-panels__panel col-lg-7" />
            <SavedSearchesPanel className="enterprise-home-panels__panel col-lg-5" />
        </div>
    </div>
)
