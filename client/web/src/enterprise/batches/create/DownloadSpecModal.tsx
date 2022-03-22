import React from 'react'

import CloseIcon from 'mdi-react/CloseIcon'

import { Button, Modal } from '@sourcegraph/wildcard'

import { CodeSnippet } from '../../../../../branded/src/components/CodeSnippet'

import styles from './DownloadSpecModal.module.scss'

export interface DownloadSpecModalProps {
    onCancel: () => void
    onConfirm: () => void
}

export const DownloadSpecModal: React.FunctionComponent<DownloadSpecModalProps> = ({ onCancel, onConfirm }) => (
    <Modal onDismiss={onCancel} aria-labelledby={MODAL_LABEL_ID}>
        <>
            <div>
                <div>
                    <h3 id={MODAL_LABEL_ID}>Download specification for src-cli</h3>
                    <Button className={styles.close} onClick={onCancel}>
                        <CloseIcon className={styles.icon} />
                    </Button>
                </div>

                <div className={styles.container}>
                    <div className={styles.left}>
                        <p className="mb-4">
                            Use the Sourcegraph CLI (src) to preview the commits and changesets that your batch change
                            will make:
                        </p>

                        <CodeSnippet code="src batch preview -f Hello world" language="bash" className={styles.test2} />

                        <p className="mb-4">
                            Follow the URL printed in your terminal to see the preview and (when you're ready) create
                            the batch change.
                        </p>
                    </div>
                    <div className={styles.right}>
                        <p className="mb-4">About src-cli </p>
                        <p className={styles.test}>
                            src cli is a command line interface to Sourcegraph. Its batch command allows to run batch
                            specification files locally.
                        </p>
                        <p className="mb-4">Download src-cli</p>
                    </div>
                </div>
                <div className="d-flex justify-content-between">
                    <div>
                        {/* <div className="d-flex justify-content-start"> */}
                        <Button
                            className="border-0 ml-2"
                            // onClick={onDismissClick}
                            outline={true}
                            variant="secondary"
                            size="sm"
                        >
                            Don't show this again
                        </Button>
                    </div>
                    {/* <div> */}
                    <div>
                        {/* <div className="d-flex justify-content-end"> */}
                        <Button className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                            Cancel
                        </Button>
                        {/* <BatchSpecDownloadLink name={batchChange.name} originalInput={code} isLightTheme={isLightTheme}> */}
                        <Button onClick={onConfirm} variant="primary">
                            Download Spec
                        </Button>
                        {/* </BatchSpecDownloadLink> */}
                    </div>
                </div>
            </div>
        </>
    </Modal>
)

const MODAL_LABEL_ID = 'download-spec-modal'
