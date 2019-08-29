import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import React, { useCallback, useState } from 'react'

/**
 * A button to copy a link to the current thread to the clipboard.
 */
export const CopyThreadLinkButton: React.FunctionComponent<{
    link: string
    className?: string
}> = ({ link, className = '', children }) => {
    const [justCopied, setJustCopied] = useState(false)
    const onCopyClick = useCallback<React.MouseEventHandler>(
        e => {
            e.preventDefault()
            copy(link)
            setJustCopied(true)
            setTimeout(() => setJustCopied(false), 1000)
        },
        [link]
    )

    return (
        <a className={className} data-tooltip={justCopied ? 'Copied!' : 'Copy link'} onClick={onCopyClick} href={link}>
            {children} <ContentCopyIcon className="icon-inline ml-1" />
        </a>
    )
}
