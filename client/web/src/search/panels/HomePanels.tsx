import * as React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { HomePanelsProps, PatternTypeProps } from '..'
import { AuthenticatedUser } from '../../auth'

import { RecentFilesPanel } from './RecentFilesPanel'
import { RecentSearchesPanel } from './RecentSearchesPanel'
import { RepogroupPanel } from './RepogroupPanel'
import { RepositoriesPanel } from './RepositoriesPanel'
import { SavedSearchesPanel } from './SavedSearchesPanel'

interface Props extends Pick<PatternTypeProps, 'patternType'>, TelemetryProps, HomePanelsProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}

export const HomePanels: React.FunctionComponent<Props> = (props: Props) => (
    <div className="home-panels container">
        <div className="row">
            <RepositoriesPanel {...props} className="home-panels__panel col-lg-4" />
            <RecentSearchesPanel {...props} className="home-panels__panel col-lg-8" />
        </div>
        <div className="row">
            <RecentFilesPanel {...props} className="home-panels__panel col-lg-7" />

            {props.isSourcegraphDotCom ? (
                <RepogroupPanel {...props} className="home-panels__panel col-lg-5" />
            ) : (
                <SavedSearchesPanel {...props} className="home-panels__panel col-lg-5" />
            )}
        </div>
    </div>
)
