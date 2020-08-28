import * as React from 'react'
import { RecentFilesPanel } from './RecentFilesPanel'
import { RecentSearchesPanel } from './RecentSearchesPanel'
import { RepositoriesPanel } from './RepositoriesPanel'
import { SavedSearchesPanel } from './SavedSearchesPanel'

export const EnterpriseHomePanels: React.FunctionComponent<{}> = () => (
    <>
        <RepositoriesPanel />
        <RecentSearchesPanel />
        <RecentFilesPanel />
        <SavedSearchesPanel />
    </>
)
