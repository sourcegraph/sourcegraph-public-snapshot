import React from 'react'

import classNames from 'classnames'

import { Alert, Button, H2, LoadingSpinner, Modal, Text } from '@sourcegraph/wildcard'

import { UpdateInfo } from './updater'

export interface InstallModalProps {
    details: UpdateInfo
    onClose: () => void
}

const labelId = 'newVersionInfo'

export const InstallModal: React.FC<InstallModalProps> = ({ details, onClose }) => (
    <Modal className={classNames('d-flex flex-column')} aria-labelledby={labelId} style={{ maxHeight: '60%' }}>
        <H2 className="mb-4">Installing Version {details.newVersion}</H2>
        {['INSTALLING', 'PENDING', 'IDLE'].includes(details.stage) && (
            <>
                <div className="d-flex p-2 mt-2 flex-grow-1">
                    <LoadingSpinner />
                    <Text className="ml-2 mb-1">Please wait. The app will restart after upgrade.</Text>
                </div>
            </>
        )}
        {details.stage === 'ERROR' && (
            <div className="mt-4">
                <Alert variant="danger">{details.error}</Alert>
            </div>
        )}
        <div className="d-flex justify-content-end">
            {details.stage === 'ERROR' ? (
                <>
                    <Button variant="primary" className="m-1" onClick={details.startInstall}>
                        Retry
                    </Button>
                    <Button variant="secondary" className="m-1" onClick={onClose}>
                        Close
                    </Button>
                </>
            ) : (
                <Button variant="secondary" className="m-1" onClick={onClose}>
                    Cancel
                </Button>
            )}
        </div>
    </Modal>
)
