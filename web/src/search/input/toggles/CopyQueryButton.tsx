import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import React, { useCallback } from 'react'
import classNames from 'classnames'
import { Observable, merge, of, timer } from 'rxjs'
import { tap, switchMapTo, startWith } from 'rxjs/operators'
import { useEventObservable } from '../../../../../shared/src/util/useObservable'

interface Props {
    fullQuery: string
    className?: string
}

/**
 * A repository header action that copies the current page's URL to the clipboard.
 */
export const CopyQueryButton: React.FunctionComponent<Props> = (props: Props) => {
    const [nextClick, copied] = useEventObservable(
        useCallback(
            (clicks: Observable<React.MouseEvent>) =>
                clicks.pipe(
                    tap(() => copy(props.fullQuery)),
                    switchMapTo(merge(of(true), timer(1000).pipe(switchMapTo(of(false))))),
                    tap(() => Tooltip.forceUpdate()),
                    startWith(false)
                ),
            [props.fullQuery]
        )
    )

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
