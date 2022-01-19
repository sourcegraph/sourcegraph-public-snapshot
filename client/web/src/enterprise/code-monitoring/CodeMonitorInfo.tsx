import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import React from 'react'

import { Button } from '@sourcegraph/wildcard'

import { useTemporarySetting } from '../../settings/temporary/useTemporarySetting'

import styles from './CodeMonitorInfo.module.scss'

export const CodeMonitorInfo: React.FunctionComponent<{ className?: string }> = React.memo(({ className }) => {
    const [visible, setVisible] = useTemporarySetting('codemonitor.info.visible')

    if (visible === false) {
        return null
    }

    return (
        <div className={classNames('alert alert-info alert-dismissable d-flex align-items-start', className)}>
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
                className={classNames('btn-icon', styles.closeButton)}
                onClick={() => setVisible(false)}
            >
                <CloseIcon className="icon-inline" />
            </Button>
        </div>
    )
})
