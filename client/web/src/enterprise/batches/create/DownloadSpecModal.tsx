import React from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
import CloseIcon from 'mdi-react/CloseIcon'

import { Button, Link, Modal } from '@sourcegraph/wildcard'

import { CodeSnippet } from '../../../../../branded/src/components/CodeSnippet'
import { BatchSpecDownloadLink, getFileName } from '../BatchSpec'

import styles from './DownloadSpecModal.module.scss'

export interface DownloadSpecModalProps {
    name: string
    originalInput: string
    isLightTheme: boolean
    setIsDownloadSpecModalOpen: (condition: boolean) => void
    setDownloadSpecModalDismissed: (condition: boolean) => void
}

export const DownloadSpecModal: React.FunctionComponent<DownloadSpecModalProps> = ({
    name,
    originalInput,
    isLightTheme,
    setIsDownloadSpecModalOpen,
    setDownloadSpecModalDismissed,
}) => (
    <Modal
        onDismiss={() => {
            setIsDownloadSpecModalOpen(false)
        }}
        aria-labelledby={MODAL_LABEL_ID}
        className={styles.modal}
    >
        <div>
            <h3 id={MODAL_LABEL_ID}>Download spec for src-cli</h3>
            <Button
                className={styles.close}
                onClick={() => {
                    setIsDownloadSpecModalOpen(false)
                }}
            >
                <VisuallyHidden>Close</VisuallyHidden>
                <CloseIcon className={styles.icon} />
            </Button>
        </div>

        <div className={styles.container}>
            <div className={styles.left}>
                <p>
                    Use the <Link to="https://docs.sourcegraph.com/cli">Sourcegraph CLI (src) </Link>
                    to run this batch change locally.
                </p>

                <CodeSnippet
                    code={`src batch preview -f ${getFileName(name)}`}
                    language="bash"
                    className={styles.codeSnippet}
                />

                <p className="p-0 m-0">
                    Follow the URL printed in your terminal to see the preview and (when you're ready) create the batch
                    change.
                </p>
            </div>
            <div className={styles.right}>
                <div>
                    <h4>About src-cli </h4>
                    <p>
                        src cli is a command line interface to Sourcegraph. Its{' '}
                        <span className="text-monospace">batch</span> command allows to run batch specification files
                        using Docker.
                    </p>
                    <Link to="https://docs.sourcegraph.com/cli">Download src-cli</Link>
                </div>
            </div>
        </div>
        <div className="d-flex justify-content-between">
            <Button className="p-0" onClick={() => setDownloadSpecModalDismissed(true)} variant="link">
                Don't show this again
            </Button>
            <div>
                <Button
                    className="mr-2"
                    outline={true}
                    variant="secondary"
                    onClick={() => {
                        setIsDownloadSpecModalOpen(false)
                    }}
                >
                    Cancel
                </Button>
                <BatchSpecDownloadLink name={name} originalInput={originalInput} isLightTheme={isLightTheme}>
                    <Button
                        variant="primary"
                        onClick={() => {
                            setIsDownloadSpecModalOpen(false)
                        }}
                    >
                        Download spec
                    </Button>
                </BatchSpecDownloadLink>
            </div>
        </div>
    </Modal>
)

const MODAL_LABEL_ID = 'download-spec-modal'
