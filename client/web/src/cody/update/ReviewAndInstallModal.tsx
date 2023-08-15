import React, { useEffect, useState } from 'react'

import classNames from 'classnames'

import { renderMarkdown } from '@sourcegraph/common'
import { Alert, Button, H2, LoadingSpinner, Markdown, Modal, Text } from '@sourcegraph/wildcard'

import type { UpdateInfo } from './updater'

import styles from './ReviewAndInstallModal.module.scss'

export interface ChangelogModalProps {
    fromSettingsPage?: boolean
    details: UpdateInfo
    onClose: () => void
}

const labelId = 'newVersionInfo'

export const ChangelogModal: React.FC<ChangelogModalProps> = ({ details, fromSettingsPage = false, onClose }) => {
    const [installing, setInstalling] = useState<boolean>(false)
    const [failed, setFailed] = useState<boolean>(false)

    useEffect(() => {
        if (details.stage === 'ERROR') {
            setFailed(true)
            setInstalling(false)
        }
    }, [details.stage])

    return (
        <Modal
            className={classNames('d-flex flex-column', styles.modal)}
            aria-labelledby={labelId}
            style={{ maxHeight: '60%' }}
        >
            <H2 className="mb-4">New Version Available: {details.newVersion}</H2>
            <div className={classNames('p-3 overflow-auto', styles.info)}>
                {details.description ? (
                    <Markdown dangerousInnerHTML={renderMarkdown(details.description)} />
                ) : (
                    <LoadingSpinner />
                )}
            </div>
            {details.stage === 'ERROR' && (
                <div className="mt-4">
                    <Alert variant="danger">
                        {details.error}
                        {fromSettingsPage ? (
                            details.checkNow !== undefined && (
                                <Button
                                    variant="link"
                                    onClick={() => {
                                        details.checkNow?.(true)
                                    }}
                                >
                                    Try Again
                                </Button>
                            )
                        ) : (
                            <Text className="mt-2">Please visit Settings &gt; About to retry install.</Text>
                        )}
                    </Alert>
                </div>
            )}
            <div className="d-flex justify-content-end mt-4">
                {['INSTALLING', 'PENDING'].includes(details.stage) && (
                    <div className="d-flex p-2 mt-2 flex-grow-1">
                        <LoadingSpinner />
                        <Text className="ml-2 mb-1">Installing... Please wait...</Text>
                    </div>
                )}
                <Button
                    className="m-1 mt-2"
                    variant="primary"
                    onClick={() => {
                        setInstalling(true)
                        details.startInstall?.()
                    }}
                    disabled={installing || failed}
                >
                    Update and Restart
                </Button>
                <Button className="m-1 mt-2" variant="secondary" onClick={onClose} disabled={installing}>
                    Cancel
                </Button>
            </div>
        </Modal>
    )
}
