import React from 'react'

import { FileLinkProps } from '@sourcegraph/cody-ui/src/chat/ContextFiles'

import { getVSCodeAPI } from './utils/VSCodeApi'

import styles from './FileLink.module.css'

export const FileLink: React.FunctionComponent<FileLinkProps> = ({ path }) => (
    <button
        className={styles.linkButton}
        type="button"
        onClick={() => {
            getVSCodeAPI().postMessage({ command: 'openFile', filePath: path })
        }}
    >
        {path}
    </button>
)
