import React from 'react'

import classNames from 'classnames'

import styles from './LogOutput.module.scss'

export interface LogOutputProps {
    text: string
    className?: string
}

export const LogOutput: React.FunctionComponent<React.PropsWithChildren<LogOutputProps>> = React.memo(
    ({ text, className }) => (
        <pre className={classNames(styles.logs, 'rounded p-3 mb-0', className)}>
            {
                // Use index as key because log lines may not be unique. This is OK
                // here because this list will not be updated during this component's
                // lifetime (note: it's also memoized).
                /* eslint-disable react/no-array-index-key */
                text.split('\n').map((line, index) => (
                    <code
                        key={index}
                        className={classNames('d-block', line.startsWith('stderr:') ? 'text-danger' : '')}
                    >
                        {line.replace(/^std(out|err): /, '')}
                    </code>
                ))
            }
        </pre>
    )
)
