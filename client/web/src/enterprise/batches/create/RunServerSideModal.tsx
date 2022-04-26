import React from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
import CloseIcon from 'mdi-react/CloseIcon'

import { Button, Link, Modal } from '@sourcegraph/wildcard'

import styles from './RunServerSideModal.module.scss'

export interface RunServerSideModalProps {
    setIsRunServerSideModalOpen: (condition: boolean) => void
}

export const RunServerSideModal: React.FunctionComponent<RunServerSideModalProps> = ({
    setIsRunServerSideModalOpen,
}) => (
    <Modal
        onDismiss={() => {
            setIsRunServerSideModalOpen(false)
        }}
        aria-labelledby={MODAL_LABEL_ID}
        className={styles.modal}
    >
        <div>
            <h3 id={MODAL_LABEL_ID}>Running batch changes server-side is not enabled</h3>
            <Button
                className={styles.close}
                onClick={() => {
                    setIsRunServerSideModalOpen(false)
                }}
            >
                <VisuallyHidden>Close</VisuallyHidden>
                <CloseIcon className={styles.icon} />
            </Button>
        </div>

        <div className={styles.content}>
            <div className={styles.left}>
                <div>
                    <p>
                        Install executors to enable running batch changes server-side instead of locally. Executors can
                        also be autoscaled to speed up creating large-scale batch changes.
                    </p>
                </div>

                <div className={styles.videoContainer}>Video</div>
            </div>
            <div className={styles.right}>
                <div className={styles.rightTop}>
                    <h4>Resources</h4>
                    <ul>
                        <Link to="https://docs.sourcegraph.com/batch_changes/explanations/server_side">
                            <li>Running batch changes server-side</li>
                        </Link>
                        <Link to="https://docs.sourcegraph.com/admin/executors">
                            <li>Deploying executors</li>
                        </Link>
                    </ul>
                </div>

                <div className={styles.rightBottom}>
                    <div className={styles.blank}>
                        <h4>Request a demo</h4>
                        <p>Learn more about this free feature of batch changes.</p>

                        <Button variant="primary">Request Demo</Button>
                    </div>
                </div>
            </div>
        </div>
    </Modal>
)

const MODAL_LABEL_ID = 'run-server-side-modal'
