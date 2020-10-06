import classNames from 'classnames'
import React, { useCallback } from 'react'
import { AuthenticatedUser } from '../../auth'
import { EventLogResult } from '../backend'
import { Link } from '../../../../shared/src/components/Link'
import { Observable } from 'rxjs'
import { PanelContainer } from './PanelContainer'
import { repogroupList } from '../../repogroups/HomepageConfig'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'

interface Props extends TelemetryProps {
    className?: string
    authenticatedUser: AuthenticatedUser | null
    fetchRecentFileViews: (userId: string, first: number) => Observable<EventLogResult | null>
}

export const RepogroupPanel: React.FunctionComponent<Props> = ({ className, telemetryService }) => {
    const logRepogroupClicked = useCallback(() => telemetryService.log('RepogroupPanelRepogroupClicked'), [
        telemetryService,
    ])

    const populatedContent = (
        <>
            {repogroupList.map(repogroup => (
                <div className="d-flex mt-2" key={repogroup.name}>
                    <img className="repogroup-panel__repogroup-icon mr-2" src={repogroup.homepageIcon} />
                    <div className="d-flex flex-column">
                        <Link to={repogroup.url} onClick={logRepogroupClicked} className="mb-1 font-weight-bold">
                            {repogroup.title}
                        </Link>
                        <p>{repogroup.homepageDescription}</p>
                    </div>
                </div>
            ))}
        </>
    )

    return (
        <PanelContainer
            className={classNames(className, 'repogroup-panel')}
            title="Repository groups"
            state="populated"
            loadingContent={<></>}
            populatedContent={populatedContent}
            emptyContent={<></>}
        />
    )
}
