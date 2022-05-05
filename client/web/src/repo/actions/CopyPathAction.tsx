import React, { useState, useLayoutEffect } from 'react'

import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import { useLocation } from 'react-router'

import { Button, TooltipController, Icon } from '@sourcegraph/wildcard'

import { eventLogger } from '../../tracking/eventLogger'
import { parseBrowserRepoURL } from '../../util/url'

import styles from './CopyPathAction.module.scss'

/**
 * A repository header action that copies the current page's repository or file path to the clipboard.
 */
export const CopyPathAction: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => {
    const location = useLocation()
    const [copied, setCopied] = useState(false)

    useLayoutEffect(() => {
        TooltipController.forceUpdate()
    }, [copied])

    const onClick = (event: React.MouseEvent<HTMLButtonElement>): void => {
        event.preventDefault()
        eventLogger.log('CopyFilePath')
        const { repoName, filePath } = parseBrowserRepoURL(location.pathname)
        copy(filePath || repoName) // copy the file path if present; else it's the repo path.
        setCopied(true)

        setTimeout(() => {
            setCopied(false)
        }, 1000)
    }

    return (
        <Button
            variant="icon"
            className="p-2"
            data-tooltip={copied ? 'Copied!' : 'Copy path to clipboard'}
            aria-label="Copy path"
            onClick={onClick}
            size="sm"
        >
            <Icon className={styles.copyIcon} as={ContentCopyIcon} />
        </Button>
    )
}
