import React from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'

import { Code } from '@sourcegraph/wildcard'

import styles from './LogOutput.module.scss'

export interface LogOutputProps {
    text: string
    className?: string
    /**
     * Descriptive prefix text, visible only to screen reader.
     */
    logDescription?: string
}

export const LogOutput: React.FunctionComponent<React.PropsWithChildren<LogOutputProps>> = React.memo(
    function LogOutput({ text, className, logDescription }) {
        return (
            <>
                {logDescription && <VisuallyHidden>{logDescription}</VisuallyHidden>}
                <pre className={classNames(styles.logs, 'rounded p-3 mb-0', className)}>
                    {
                        // Use index as key because log lines may not be unique. This is OK
                        // here because this list will not be updated during this component's
                        // lifetime (note: it's also memoized).
                        /* eslint-disable react/no-array-index-key */
                        text.split('\n').map((line, index) => (
                            <Code
                                key={index}
                                className={classNames('d-block', line.startsWith('stderr:') ? 'text-danger' : '')}
                                tabIndex={0}
                                role="code"
                            >
                                {line.replace(/^std(out|err): /, '')}
                            </Code>
                        ))
                    }
                </pre>
            </>
        )
    }
)
