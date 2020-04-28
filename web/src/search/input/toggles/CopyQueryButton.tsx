import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import React, { useState, useCallback, useEffect, useMemo } from 'react'
import classNames from 'classnames'
import { Subject } from 'rxjs'
import { tap, debounceTime } from 'rxjs/operators'

interface Props {
    fullQuery: string
    className?: string
}

/**
 * A repository header action that copies the current page's URL to the clipboard.
 */
export const CopyQueryButton: React.FunctionComponent<Props> = (props: Props) => {
    const [copied, setCopied] = useState<boolean>(false)
    const click = useMemo(() => new Subject(), [])
    // use this as onClick handler
    const nextClick = useCallback(() => click.next(), [click])
    useEffect(() => {
        const sub = click
            .pipe(
                tap(() => {
                    copy(props.fullQuery)
                    setCopied(true)
                    Tooltip.forceUpdate()
                }),
                debounceTime(1000),
                tap(() => {
                    setCopied(false)
                    Tooltip.forceUpdate()
                })
            )
            .subscribe()
        return () => sub.unsubscribe()
    }, [click, props.fullQuery])

    return (
        <button
            type="button"
            className={classNames('btn btn-icon icon-inline  btn-link-sm', props.className)}
            data-tooltip={copied ? 'Copied!' : 'Copy query to clipboard'}
            onClick={nextClick}
        >
            <ContentCopyIcon className="icon-inline" />
        </button>
    )
}
