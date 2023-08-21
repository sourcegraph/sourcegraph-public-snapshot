import { UpdateInfoContent } from './UpdateInfoContent'
import { type UpdateInfo, useUpdater } from './updater'

export interface AboutPageUpdateFrameProps {
    update: UpdateInfo
}

export const AboutPageUpdateFrame: React.FC<AboutPageUpdateFrameProps> = ({ update }) => (
    <UpdateInfoContent details={update} fromSettingsPage={false} />
)

export function AboutPageUpdatePanel(): JSX.Element {
    const update = useUpdater()
    return <AboutPageUpdateFrame update={update} />
}
