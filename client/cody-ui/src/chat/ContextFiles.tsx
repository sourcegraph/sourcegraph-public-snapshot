/* eslint-disable jsx-a11y/no-static-element-interactions */
/* eslint-disable jsx-a11y/click-events-have-key-events */
import React, { useState } from 'react'

import './ContextFiles.css'

export interface FileLinkProps {
    path: string
}

export const ContextFiles: React.FunctionComponent<{
    contextFiles: string[]
    fileLinkComponent: React.FunctionComponent<FileLinkProps>
}> = ({ contextFiles, fileLinkComponent: FileLink }) => {
    const [isExpanded, setIsExpanded] = useState(false)

    if (contextFiles.length === 1) {
        return (
            <p>
                Cody read{' '}
                <code className="context-file">
                    <FileLink path={contextFiles[0]} />
                </code>{' '}
                to provide an answer.
            </p>
        )
    }

    if (isExpanded) {
        return (
            <div className="context-files-expanded">
                <span className="context-files-toggle-icon" onClick={() => setIsExpanded(false)}>
                    <i className="codicon codicon-triangle-down" slot="start" />
                </span>
                <div>
                    <div className="context-files-list-title" onClick={() => setIsExpanded(false)}>
                        Cody read the following files to provide an answer:
                    </div>
                    <ul className="context-files-list-container">
                        {contextFiles.map(file => (
                            <li key={file}>
                                <code className="context-file">
                                    <FileLink path={file} />
                                </code>
                            </li>
                        ))}
                    </ul>
                </div>
            </div>
        )
    }

    return (
        <div className="context-files-collapsed" onClick={() => setIsExpanded(true)}>
            <span className="context-files-toggle-icon">
                <i className="codicon codicon-triangle-right" slot="start" />
            </span>
            <div className="context-files-collapsed-text">
                <span>
                    Cody read <code className="context-file">{contextFiles[0].split('/').pop()}</code> and{' '}
                    {contextFiles.length - 1} other {contextFiles.length > 2 ? 'files' : 'file'} to provide an answer.
                </span>
            </div>
        </div>
    )
}
