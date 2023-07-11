import { Alert } from '@sourcegraph/wildcard'

import { UpdateInfoContent } from './UpdateInfoContent'
import { UpdateInfo, useUpdater } from './updater'

export interface HomePageUpdateNoticeFrameProps {
    update: UpdateInfo
}

export const HomePageUpdateNoticeFrame: React.FC<HomePageUpdateNoticeFrameProps> = ({ update }) =>
    update.hasNewVersion ? (
        <div className="p-1">
            <Alert variant="info">
                <UpdateInfoContent details={update} />
            </Alert>
        </div>
    ) : (
        <></>
    )

export function HomePageUpdateNotice(): JSX.Element {
    const update = useUpdater()
    return <HomePageUpdateNoticeFrame update={update} />
}
