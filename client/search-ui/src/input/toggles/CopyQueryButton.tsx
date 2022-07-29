import React, { useCallback, useRef, useEffect } from 'react'

import { mdiClipboardOutline } from '@mdi/js'
import VisuallyHidden from '@reach/visually-hidden'
import { Shortcut } from '@slimsag/react-shortcuts'
import classNames from 'classnames'
import copy from 'copy-to-clipboard'
import { Observable, merge, of } from 'rxjs'
import { tap, switchMapTo, startWith, delay } from 'rxjs/operators'

import { useKeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts/useKeyboardShortcut'
import { Button, Icon, DeprecatedTooltipController, useEventObservable, Tooltip } from '@sourcegraph/wildcard'

interface Props {
    fullQuery: string
    className?: string
    isMacPlatform: boolean
}

/**
 * A repository header action that copies the current page's URL to the clipboard.
 */
export const CopyQueryButton: React.FunctionComponent<React.PropsWithChildren<Props>> = (props: Props) => {
    // Convoluted, but using props.fullQuery directly in the copyFullQuery callback does not work, since
    // props.fullQuery is not refrenced during the render and it is not updated within the callback.
    const fullQueryReference = useRef<string>('')
    useEffect(() => {
        fullQueryReference.current = props.fullQuery
    }, [props.fullQuery])

    const copyFullQuery = useCallback((): boolean => copy(fullQueryReference.current), [fullQueryReference])

    const [nextClick, copied] = useEventObservable(
        useCallback(
            (clicks: Observable<React.MouseEvent>) =>
                clicks.pipe(
                    tap(copyFullQuery),
                    switchMapTo(merge(of(true), of(false).pipe(delay(2000)))),
                    tap(() => DeprecatedTooltipController.forceUpdate()),
                    startWith(false)
                ),
            [copyFullQuery]
        )
    )

    const fullCopyShortcut = useKeyboardShortcut('copyFullQuery')

    const copyFullQueryTooltip = `Copy full query\n${props.isMacPlatform ? '⌘' : 'Ctrl'}+⇧+C`
    return (
        <>
            {copied && <VisuallyHidden aria-live="polite">Copied!</VisuallyHidden>}
            <Tooltip content={copied ? 'Copied!' : copyFullQueryTooltip}>
                <Button
                    className={classNames('btn-icon', props.className)}
                    variant="icon"
                    size="sm"
                    aria-label={copyFullQueryTooltip}
                    onClick={nextClick}
                >
                    <Icon aria-hidden={true} svgPath={mdiClipboardOutline} />
                </Button>
            </Tooltip>
            {fullCopyShortcut?.keybindings.map((keybinding, index) => (
                <Shortcut key={index} {...keybinding} onMatch={copyFullQuery} allowDefault={false} ignoreInput={true} />
            ))}
        </>
    )
}
