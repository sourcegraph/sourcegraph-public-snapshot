import { useEffect, useState } from 'react'

import { ChangelogModal } from './ReviewAndInstallModal'
import { useUpdater } from './updater'

/**
 * Checks -- once -- if an update is available and displays the upgrade modal.
 *
 * Used in the main window to act as a startup reminder to upgrade.
 *
 * @returns the modal with update information and actions
 */
export function StartupUpdateChecker(): JSX.Element {
    const update = useUpdater({ keepChecking: false })
    const [firstShowing, setFirstShowing] = useState<boolean>(true)
    const [showReviewAndInstall, setShowReviewAndInstall] = useState<boolean>(false)

    useEffect(() => {
        if (update.hasNewVersion && firstShowing) {
            setShowReviewAndInstall(true)
            setFirstShowing(false)
        }
    }, [update, firstShowing])

    return showReviewAndInstall ? (
        <ChangelogModal details={update} fromSettingsPage={false} onClose={() => setShowReviewAndInstall(false)} />
    ) : (
        <></>
    )
}
