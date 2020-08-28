import * as React from 'react'
import { RecentFilesPanel } from './RecentFilesPanel'
import { RecentSearchesPanel } from './RecentSearchesPanel'
import { RepositoriesPanel } from './RepositoriesPanel'
import { SavedSearchesPanel } from './SavedSearchesPanel'

export const EnterpriseHomePanels: React.FunctionComponent<{}> = () => (
    <div className="enterprise-home-panels">
        <div className="enterprise-home-panels__row">
            <RepositoriesPanel />
            <RecentSearchesPanel />
        </div>
        <div className="enterprise-home-panels__row">
            <RecentFilesPanel />
            <SavedSearchesPanel />
        </div>
    </div>
)
