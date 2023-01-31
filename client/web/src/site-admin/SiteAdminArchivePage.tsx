import React, { ChangeEvent, useCallback, useMemo, useState } from 'react'

import classNames from 'classnames'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    PageHeader,
    Container,
    AnchorLink,
    Button,
    Input,
    Text,
    Alert,
    ProductStatusBadge,
    Modal,
    H3,
} from '@sourcegraph/wildcard'

import styles from './SiteAdminArchivePage.module.scss'

import { PageTitle } from '../components/PageTitle'

interface Props extends TelemetryProps {}

/**
 * A page for archiving and restoring a Sourcegraph instance
 */
export const SiteAdminArchivePage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({ telemetryService }) => {
    useMemo(() => {
        telemetryService.logPageView('SiteAdminArchive')
    }, [telemetryService])

    const [file, setFile] = useState<File | null>(null)
    // TODO(jac): deduplicate
    const [showSuccessModal, setShowSuccessModal] = useState<Boolean>(false)
    const [showFailureModal, setShowFailureModal] = useState<Boolean>(false)

    const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files[0]) {
            setFile(e.target.files[0])
        }
    }

    const upload = useCallback(
        (event: React.MouseEvent<HTMLButtonElement, MouseEvent>) => {
            if (!file) {
                return
            }

            event.preventDefault()

            const data = new FormData()
            data.append('archive', file, file.name)

            fetch('/site-admin/archive/upload', {
                method: 'POST',
                body: data,
            }).then(res => {
                if (res.ok) {
                    setShowSuccessModal(true)
                } else {
                    setShowFailureModal(true)
                }
            })
        },
        [file]
    )

    return (
        <div>
            <PageTitle title="Archive & Restore - Admin" />
            <PageHeader
                path={[{ text: 'Archive & Restore' }]}
                headingElement="h2"
                className="mb-3"
                description="Archive & Restore Sourcegraph instance configuration."
            />
            <ProductStatusBadge status="experimental" className="mb-2" />
            <Container className="mb-2">
                <H3>Archive</H3>
                <Text>
                    The archive includes <strong>non-regenerable</strong> data e.g. <em>site configuration</em>,{' '}
                    <em>code host configuration</em>. The archive <strong>does not</strong> include regenerable data
                    e.g. <em>cloned repositories</em>, <em>search indexes</em>.
                </Text>
                <Button to="/site-admin/archive/download" download="true" variant="primary" as={AnchorLink}>
                    Download Archive
                </Button>
            </Container>

            <Container className="mb-2">
                {/* TODO(jac): deduplicate */}
                {showSuccessModal && (
                    <Modal aria-labelledby="Restore - Success">
                        <H3 className="mb-4">Restore - Success</H3>
                        <div className="form-group mb-4">
                            Please restart the Sourcegraph instance for the changes to take effect.
                        </div>
                        <div className="d-flex justify-content-end">
                            <Button className="mr-2" onClick={() => setShowSuccessModal(false)} variant="primary">
                                Close
                            </Button>
                        </div>
                    </Modal>
                )}
                {showFailureModal && (
                    <Modal aria-labelledby="Restore - Failure">
                        <H3 className="mb-4">Restore - Failure</H3>
                        <div className="form-group mb-4">
                            Restore failed - the instance may be in a bad state. Please revert to a backup
                        </div>
                        <div className="d-flex justify-content-end">
                            <Button className="mr-2" onClick={() => setShowFailureModal(false)} variant="primary">
                                Close
                            </Button>
                        </div>
                    </Modal>
                )}
                <H3>Restore</H3>
                <Text>
                    Restore Sourcegraph instance configuration from an archive. The archive <strong>must</strong> have
                    been exported from a Sourcegraph instance running the same version as this instance.
                </Text>
                <Text>For the restore to take effect the instance must be restarted.</Text>
                <Alert variant="warning">
                    <h4>Restore is a destructive action</h4>
                    The current instance configuration will be overwritten.
                    <Text>
                        This feature is currently experimental. As such it is advised that a backup has been created in
                        the event of a restore failure.
                    </Text>
                </Alert>
                <Input
                    label="Select Archive File"
                    type="file"
                    name="archive"
                    onChange={handleFileChange}
                    className={classNames(styles.selectInput)}
                />

                <Button className="m2-4" type="submit" variant="danger" onClick={upload}>
                    Restore
                </Button>
            </Container>
        </div>
    )
}
