import classNames from 'classnames'
import React from 'react'

export const CodeMonitorInfo: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <p className={classNames('alert alert-info', className)}>
        We currently recommend code monitors on repositories that don’t have a high commit traffic and for non-critical
        use cases.
        <br />
        We are actively working on increasing the performance and fidelity of code monitors to support more sensitive
        workloads, like a large number of repositories or auditing published code for secrets and other security use
        cases.
    </p>
)
