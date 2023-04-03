import React from 'react'

import { vscodeAPI } from './utils/VSCodeApi'

import './FileLink.css'

import { FileLinkProps } from '@sourcegraph/cody-ui/src/chat/ContextFiles'

export const FileLink: React.FunctionComponent<FileLinkProps> = ({ path }) => (
    <button
        className="link-button"
        type="button"
        onClick={() => {
            vscodeAPI.postMessage({ command: 'openFile', filePath: path })
        }}
    >
        {path}
    </button>
)
