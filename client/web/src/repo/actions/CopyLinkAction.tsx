import classNames from 'classnames'
import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import React, { useState, useLayoutEffect } from 'react'
import { useLocation } from 'react-router'

import { Tooltip } from '@sourcegraph/branded/src/components/tooltip/Tooltip'
import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'

import { eventLogger } from '../../tracking/eventLogger'

import styles from './CopyLinkAction.module.scss'

/**
 * A repository header action that copies the current page's URL to the clipboard.
 */
export const CopyLinkAction: React.FunctionComponent = () => {
    const [isRedesignEnabled] = useRedesignToggle()
    const location = useLocation()
    const [copied, setCopied] = useState(false)

    useLayoutEffect(() => {
        Tooltip.forceUpdate()
    }, [copied])

    const onClick = (event: React.MouseEvent<HTMLButtonElement>): void => {
        event.preventDefault()
        eventLogger.log('ShareButtonClicked')
        const shareLink = new URL(location.pathname + location.search + location.hash, window.location.href)
        shareLink.searchParams.set('utm_source', 'share')
        copy(shareLink.href)

        setCopied(true)

        setTimeout(() => {
            setCopied(false)
        }, 1000)
    }

    return (
        <button
            type="button"
            className={classNames('btn btn-icon my-2', isRedesignEnabled && 'btn-sm')}
            data-tooltip={copied ? 'Copied!' : 'Copy link to clipboard'}
            aria-label="Copy link"
            onClick={onClick}
        >
            <ContentCopyIcon className={classNames('icon-inline', styles.copyIcon)} />
        </button>
    )
}
