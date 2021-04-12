import classNames from 'classnames'
import React, { useCallback } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../../auth'
import { repogroupList } from '../../repogroups/HomepageConfig'

import { PanelContainer } from './PanelContainer'

interface Props extends TelemetryProps {
    className?: string
    authenticatedUser: AuthenticatedUser | null
}

export const RepogroupPanel: React.FunctionComponent<Props> = ({ className, telemetryService }) => {
    const logRepogroupClicked = useCallback(() => telemetryService.log('RepogroupPanelRepogroupClicked'), [
        telemetryService,
    ])

    const populatedContent = (
        <div className="mt-2 row">
            {repogroupList.map(repogroup => (
                <div className="d-flex align-items-center mb-4 col-xl-6 col-lg-12 col-sm-6" key={repogroup.name}>
                    <img className="repogroup-panel__repogroup-icon mr-4" src={repogroup.homepageIcon} alt="" />
                    <div className="d-flex flex-column">
                        <Link to={repogroup.url} onClick={logRepogroupClicked} className="mb-1">
                            {repogroup.title}
                        </Link>
                    </div>
                </div>
            ))}
        </div>
    )

    return (
        <PanelContainer
            className={classNames(className, 'repogroup-panel')}
            title="Repository groups"
            state="populated"
            populatedContent={populatedContent}
        />
    )
}
