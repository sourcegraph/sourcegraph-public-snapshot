import React from 'react'

import { mdiClose } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'

import { Button, Link, Modal, H3, H4, Text, Icon } from '@sourcegraph/wildcard'

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
        <H3 id={MODAL_LABEL_ID}>Running batch changes server-side is not enabled</H3>
        <Button
            className={styles.close}
            onClick={() => {
                setIsRunServerSideModalOpen(false)
            }}
        >
            <VisuallyHidden>Close</VisuallyHidden>
            <Icon className={styles.icon} svgPath={mdiClose} inline={false} aria-hidden={true} />
        </Button>

        <div className={styles.content}>
            <div className={styles.left}>
                <Text>
                    Install executors to enable running batch changes server-side instead of locally. Executors can also
                    be autoscaled to speed up creating large-scale batch changes.
                </Text>

                <video
                    className="w-100 h-auto shadow percy-hide"
                    width={1280}
                    height={720}
                    autoPlay={true}
                    muted={true}
                    loop={true}
                    playsInline={true}
                    controls={false}
                >
                    <source type="video/webm" src="https://storage.googleapis.com/sourcegraph-assets/ssbc_demo.webm" />
                    <source type="video/mp4" src="https://storage.googleapis.com/sourcegraph-assets/ssbc_demo.mp4" />
                </video>
            </div>
            <div className={styles.right}>
                <div className={styles.rightTop}>
                    <H4>Resources</H4>
                    <ul className={styles.linksList}>
                        <Link to="/help/batch_changes/explanations/server_side">
                            <li>Running batch changes server-side</li>
                        </Link>
                        <Link to="/help/admin/executors">
                            <li>Deploying executors</li>
                        </Link>
                    </ul>
                </div>

                {/* TODO: Restore this once we have a process and link for requesting this demo */}
                {/* <div className={styles.rightBottom}>
                    <div className={styles.blank}>
                        <H4>Request a demo</H4>
                        <Text>Learn more about this free feature of batch changes.</Text>

                        <Button variant="primary">Request Demo</Button>
                    </div>
                </div> */}
            </div>
        </div>
    </Modal>
)

const MODAL_LABEL_ID = 'run-server-side-modal'
