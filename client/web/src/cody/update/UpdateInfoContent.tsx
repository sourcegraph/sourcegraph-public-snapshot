import { useState } from 'react'

import { mdiCheckCircle } from '@mdi/js'
import classNames from 'classnames'

import { Button, Icon, LoadingSpinner, Text } from '@sourcegraph/wildcard'

import { ChangelogModal } from './ReviewAndInstallModal'
import type { UpdateInfo } from './updater'

import styles from './UpdateInfoContent.module.scss'

interface UpdateInfoContentProps {
    details: UpdateInfo
    fromSettingsPage?: boolean
}

function UpdateDetails({ details, fromSettingsPage }: UpdateInfoContentProps): JSX.Element {
    const [reviewAndInstall, setReviewAndInstall] = useState<boolean>(false)

    return (
        <>
            <div className="d-flex align-items-center m-0">
                <Text className="m-0">
                    Cody version upgrade from {details.version} to <b>{details.newVersion}</b> update available.
                </Text>
                <Button
                    variant="link"
                    onClick={() => setReviewAndInstall(true)}
                    disabled={details.startInstall === undefined}
                >
                    Review and Install
                </Button>
            </div>

            {reviewAndInstall && (
                <ChangelogModal
                    details={details}
                    fromSettingsPage={fromSettingsPage}
                    onClose={() => {
                        setReviewAndInstall(false)
                        details.checkNow?.(true)
                    }}
                />
            )}
        </>
    )
}

/**
 * UpdateInfoContent component renders update information and actions.
 *
 * @param details - Update information object.
 * @param fromSettingsPage - Whether the component is rendered on settings page.
 * @returns - Rendered component.
 */
export function UpdateInfoContent({ details: update, fromSettingsPage }: UpdateInfoContentProps): JSX.Element {
    return update.stage === 'CHECKING' ? (
        <div className="d-flex align-items-center mt-2">
            <LoadingSpinner inline={true} />
            <Text className="ml-2 mb-0">Please wait... Checking for updates...</Text>
        </div>
    ) : update.hasNewVersion ? (
        <UpdateDetails details={update} fromSettingsPage={fromSettingsPage} />
    ) : (
        <div className="d-flex align-items-center">
            {update.checkNow !== undefined && (
                <Button
                    className="mr-2 p-0"
                    variant="link"
                    onClick={() => {
                        update.checkNow?.(true)
                    }}
                >
                    Check for updates
                </Button>
            )}
            <div className={classNames('d-flex align-items-center', styles.upToDate)}>
                <Icon inline={true} svgPath={mdiCheckCircle} aria-hidden={true} />
                <Text className="ml-2 mb-0">You are running the latest version.</Text>
            </div>
        </div>
    )
}
