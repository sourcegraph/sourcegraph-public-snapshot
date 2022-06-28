import React, { useCallback, useEffect } from 'react'

import classNames from 'classnames'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { communitySearchContextsList } from '../../communitySearchContexts/HomepageConfig'

import { PanelContainer } from './PanelContainer'

import styles from './CommunitySearchContextPanel.module.scss'

interface Props extends TelemetryProps {
    className?: string
    authenticatedUser: AuthenticatedUser | null
    insideTabPanel?: boolean
}

export const CommunitySearchContextsPanel: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    className,
    telemetryService,
    insideTabPanel,
}) => {
    useEffect(() => {
        telemetryService.log('HomepageContextsPanelViewed')
    }, [telemetryService])
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
            insideTabPanel={insideTabPanel}
            className={classNames(className, 'community-search-context-panel')}
            title="Community search contexts"
            state="populated"
            populatedContent={populatedContent}
        />
    )
}
