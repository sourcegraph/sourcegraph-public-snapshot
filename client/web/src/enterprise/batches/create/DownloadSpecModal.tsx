import React from 'react'

import CloseIcon from 'mdi-react/CloseIcon'

import { Button, Link, Modal } from '@sourcegraph/wildcard'

import { CodeSnippet } from '../../../../../branded/src/components/CodeSnippet'
import { BatchSpecDownloadLink } from '../BatchSpec'

import styles from './DownloadSpecModal.module.scss'

export interface DownloadSpecModalProps {
    onCancel: () => void
    onConfirm: () => void
    name: string
    originalInput: string
    isLightTheme: boolean
    showDownloadSpecModal: () => void
    setShowDownloadSpecModal: () => void
    setModal: () => void
}

export const DownloadSpecModal: React.FunctionComponent<DownloadSpecModalProps> = ({
    onCancel,
    onConfirm,
    name,
    originalInput,
    isLightTheme,
    setModal,
    showDownloadSpecModal,
    setShowDownloadSpecModal,
}) => {
    console.log('test')

    return (
        <Modal onDismiss={onCancel} aria-labelledby={MODAL_LABEL_ID} className={styles.modal}>
            <>
                <div>
                    <div>
                        <h4 id={MODAL_LABEL_ID}>Download specification for src-cli</h4>
                        <Button
                            className={styles.close}
                            onClick={() => {
                                setModal(false)
                            }}
                        >
                            <CloseIcon className={styles.icon} />
                        </Button>
                    </div>

                    <div className={styles.container}>
                        <div className={styles.left}>
                            <p>
                                Use the{' '}
                                <Link
                                    to="https://docs.sourcegraph.com/cli
"
                                >
                                    Sourcegraph CLI (src){' '}
                                </Link>
                                to run this batch change locally.
                            </p>

                            <CodeSnippet
                                code="src batch preview -f Hello world"
                                language="bash"
                                className={styles.codeSnippet}
                            />

                            <span>
                                Follow the URL printed in your terminal to see the preview and (when you're ready)
                                create the batch change.
                            </span>
                        </div>
                        <div className={styles.right}>
                            <div className={styles.rightContent}>
                                <p>About src-cli </p>
                                <p>
                                    src cli is a command line interface to Sourcegraph. Its batch command allows to run
                                    batch specification files using Docker.
                                </p>
                                <Link to="/">Download src-cli</Link>
                            </div>
                        </div>
                    </div>
                    <div className="d-flex justify-content-between">
                        <div>
                            <Button
                                className={styles.button}
                                onClick={() => setShowDownloadSpecModal(false)}
                                outline={true}
                                variant="secondary"
                                size="sm"
                            >
                                Don't show this again
                            </Button>
                        </div>
                        <div>
                            <Button
                                className="mr-2"
                                outline={true}
                                variant="secondary"
                                size="sm"
                                onClick={() => {
                                    setModal(false)
                                }}
                            >
                                Cancel
                            </Button>
                            <BatchSpecDownloadLink
                                name={name}
                                originalInput={originalInput}
                                isLightTheme={isLightTheme}
                            >
                                <Button onClick={onConfirm} variant="primary" size="sm">
                                    Download Spec
                                </Button>
                            </BatchSpecDownloadLink>
                        </div>
                    </div>
                </div>
            </>
        </Modal>
    )
}

const MODAL_LABEL_ID = 'download-spec-modal'
