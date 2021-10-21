import classNames from 'classnames'
import React, { useCallback } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../../auth'
import { communitySearchContextsList } from '../../communitySearchContexts/HomepageConfig'

import styles from './CommunitySearchContextPanel.module.scss'
import { PanelContainer } from './PanelContainer'

interface Props extends TelemetryProps {
    className?: string
    authenticatedUser: AuthenticatedUser | null
}

export const CommunitySearchContextsPanel: React.FunctionComponent<Props> = ({ className, telemetryService }) => {
    const logContextClicked = useCallback(
        () => telemetryService.log('CommunitySearchContextsPanelCommunitySearchContextClicked'),
        [telemetryService]
    )

    const populatedContent = (
        <div className="mt-2 row">
            {communitySearchContextsList.map(communitySearchContext => (
                <div
                    className="d-flex align-items-center mb-4 col-xl-6 col-lg-12 col-sm-6"
                    key={communitySearchContext.spec}
                >
                    <img className={classNames('mr-4', styles.icon)} src={communitySearchContext.homepageIcon} alt="" />
                    <div className="d-flex flex-column">
                        <Link to={communitySearchContext.url} onClick={logContextClicked} className="mb-1">
                            {communitySearchContext.title}
                        </Link>
                    </div>
                </div>
            ))}
        </div>
    )

    return (
        <PanelContainer
            className={classNames(className, 'community-search-context-panel')}
            title="Community search contexts"
            state="populated"
            populatedContent={populatedContent}
        />
    )
}
