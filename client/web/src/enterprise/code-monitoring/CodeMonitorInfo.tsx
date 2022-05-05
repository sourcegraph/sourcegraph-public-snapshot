import React from 'react'

import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { Button, Alert, Icon } from '@sourcegraph/wildcard'

import styles from './CodeMonitorInfo.module.scss'

export const CodeMonitorInfo: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = React.memo(
    ({ className }) => {
        const [visible, setVisible] = useTemporarySetting('codemonitor.info.visible')

        if (visible === false) {
            return null
        }

        return (
            <Alert className={classNames('d-flex align-items-start', className)} variant="info">
                <p className="mb-0">
                    We currently recommend code monitors on repositories that don’t have a high commit traffic and for
                    non-critical use cases.
                    <br />
                    We are actively working on increasing the performance and fidelity of code monitors to support more
                    sensitive workloads, like a large number of repositories or auditing published code for secrets and
                    other security use cases.
                </p>
                <Button
                    aria-label="Close alert"
                    variant="icon"
                    className={styles.closeButton}
                    onClick={() => setVisible(false)}
                >
                    <Icon as={CloseIcon} />
                </Button>
            </Alert>
        )
    }
)
