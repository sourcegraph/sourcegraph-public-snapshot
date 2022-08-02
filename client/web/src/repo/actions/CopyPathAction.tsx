import React, { useState } from 'react'

import { mdiContentCopy } from '@mdi/js'
import copy from 'copy-to-clipboard'
import { useLocation } from 'react-router'

import { Button, Icon, screenReaderAnnounce, Tooltip } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'
import { parseBrowserRepoURL } from '../../util/url'

import styles from './CopyPathAction.module.scss'

/**
 * A repository header action that copies the current page's repository or file path to the clipboard.
 */
export const CopyPathAction: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => {
    const location = useLocation()
    const [copied, setCopied] = useState(false)

    const onClick = (event: React.MouseEvent<HTMLButtonElement>): void => {
        event.preventDefault()
        eventLogger.log('CopyFilePath')
        const { repoName, filePath } = parseBrowserRepoURL(location.pathname)
        copy(filePath || repoName) // copy the file path if present; else it's the repo path.
        setCopied(true)
        screenReaderAnnounce('Path copied to clipboard')

        setTimeout(() => {
            setCopied(false)
        }, 1000)
    }

    const label = copied ? 'Copied!' : 'Copy path to clipboard'

    return (
        <Tooltip content={label}>
            <Button aria-label="Copy" variant="icon" className="p-2" onClick={onClick} size="sm">
                <Icon className={styles.copyIcon} aria-hidden={true} svgPath={mdiContentCopy} />
            </Button>
        </Tooltip>
    )
}
