import React from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
import CloseIcon from 'mdi-react/CloseIcon'

import { Button, Link, Modal, Typography } from '@sourcegraph/wildcard'

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
        <Typography.H3 id={MODAL_LABEL_ID}>Running batch changes server-side is not enabled</Typography.H3>
        <Button
            className={styles.close}
            onClick={() => {
                setIsRunServerSideModalOpen(false)
            }}
        >
            <VisuallyHidden>Close</VisuallyHidden>
            <CloseIcon className={styles.icon} />
        </Button>

        <div className={styles.content}>
            <div className={styles.left}>
                <p>
                    Install executors to enable running batch changes server-side instead of locally. Executors can also
                    be autoscaled to speed up creating large-scale batch changes.
                </p>

                <div className={styles.videoContainer}>Video</div>
            </div>
            <div className={styles.right}>
                <div className={styles.rightTop}>
                    <Typography.H4>Resources</Typography.H4>
                    <ul className={styles.linksList}>
                        <Link to="https://docs.sourcegraph.com/batch_changes/explanations/server_side">
                            <li>Running batch changes server-side</li>
                        </Link>
                        <Link to="https://docs.sourcegraph.com/admin/executors">
                            <li>Deploying executors</li>
                        </Link>
                    </ul>
                </div>

                {/* TODO: Restore this once we have a process and link for requesting this demo */}
                {/* <div className={styles.rightBottom}>
                    <div className={styles.blank}>
                        <Typography.H4>Request a demo</Typography.H4>
                        <p>Learn more about this free feature of batch changes.</p>

                        <Button variant="primary">Request Demo</Button>
                    </div>
                </div> */}
            </div>
        </div>
    </Modal>
)

const MODAL_LABEL_ID = 'run-server-side-modal'
