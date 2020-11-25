import React, { FunctionComponent } from 'react'
import { LsifIndexFields } from '../../../graphql-operations'

export interface ExecutorLogsProps {
    index: LsifIndexFields
    className?: string
}

export const ExecutorLogs: FunctionComponent<ExecutorLogsProps> = ({ index, className }) => (
    <>
        <h3>Output logs</h3>

        <div className={className}>
            {index.logContents && (
                <pre className="bg-code rounded p-3">
                    <code>{index.logContents}</code>
                </pre>
            )}
        </div>
    </>
)
