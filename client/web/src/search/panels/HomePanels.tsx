import classNames from 'classnames'
import * as React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { HomePanelsProps, PatternTypeProps } from '..'
import { AuthenticatedUser } from '../../auth'

import { CommunitySearchContextsPanel } from './CommunitySearchContextPanel'
import styles from './HomePanels.module.scss'
import { RecentFilesPanel } from './RecentFilesPanel'
import { RecentSearchesPanel } from './RecentSearchesPanel'
import { RepositoriesPanel } from './RepositoriesPanel'
import { SavedSearchesPanel } from './SavedSearchesPanel'

interface Props extends Pick<PatternTypeProps, 'patternType'>, TelemetryProps, HomePanelsProps {
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean
}

export const HomePanels: React.FunctionComponent<Props> = (props: Props) => (
    <div className={classNames('container', styles.homePanels)} data-testid="home-panels">
        <div className="row">
            <RepositoriesPanel {...props} className={classNames('col-lg-4', styles.panel)} />
            <RecentSearchesPanel {...props} className={classNames('col-lg-8', styles.panel)} />
        </div>
        <div className="row">
            <RecentFilesPanel {...props} className={classNames('col-lg-7', styles.panel)} />

            {props.isSourcegraphDotCom ? (
                <CommunitySearchContextsPanel {...props} className={classNames('col-lg-5', styles.panel)} />
            ) : (
                <SavedSearchesPanel {...props} className={classNames('col-lg-5', styles.panel)} />
            )}
        </div>
    </div>
)
