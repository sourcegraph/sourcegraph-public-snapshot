import { UpdateInfoContent } from './UpdateInfoContent'
import { UpdateInfo, useUpdater } from './updater'

export interface SiteAdminUpdatePanelProps {
    update: UpdateInfo
}

export const SiteAdminUpdatePanelFrame: React.FC<SiteAdminUpdatePanelProps> = ({ update }) => (
    <UpdateInfoContent details={update} />
)

export function SiteAdminUpdatePanel(): JSX.Element {
    const update = useUpdater()
    return <SiteAdminUpdatePanelFrame update={update} />
}
