import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import React, { useState, useCallback } from 'react'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import classNames from 'classnames'

interface Props {
    fullQuery: string
    className?: string
}

/**
 * A repository header action that copies the current page's URL to the clipboard.
 */
export const CopyQueryButton: React.FunctionComponent<Props> = (props: Props) => {
    const [copied, setCopied] = useState<boolean>(false)
    const onClick = useCallback(
        (event: React.MouseEvent) => {
            event.preventDefault()
            copy(props.fullQuery)
            setCopied(true)
            Tooltip.forceUpdate()

            setTimeout(() => {
                setCopied(false)
                Tooltip.forceUpdate()
            }, 1000)
        },
        [props.fullQuery]
    )

    return (
        <button
            type="button"
            className={classNames('btn btn-icon icon-inline  btn-link-sm', props.className)}
            data-tooltip={copied ? 'Copied!' : 'Copy query to clipboard'}
            onClick={onClick}
        >
            <ContentCopyIcon className="icon-inline" />
        </button>
    )
}
