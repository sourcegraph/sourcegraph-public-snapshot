import classNames from 'classnames'
import React, { useCallback } from 'react'
import * as H from 'history'
import { isInstanceOf, anyOf } from '../util/types'

interface Props {
    wrapper?: 'div' | 'span'
    dangerousInnerHTML: string
    history: H.History
    className?: string
    /** A function to attain a reference to the top-level div from a parent component. */
    refFn?: (ref: HTMLElement | null) => void
}

export const Markdown: React.FunctionComponent<Props> = ({
    wrapper: RootComponent = 'div',
    refFn,
    className,
    history,
    dangerousInnerHTML,
}: Props) => {
    // Links in markdown cannot use react-router's <Link>.
    // Prevent hitting the backend (full page reloads) for links that stay inside the app.
    const onClick = useCallback<React.MouseEventHandler<unknown>>(
        event => {
            // Do nothing if the link was requested to open in a new tab
            if (event.ctrlKey || event.metaKey) {
                return
            }
            // Check if click happened within an anchor inside the markdown
            const anchor = event.nativeEvent
                .composedPath()
                .slice(0, event.nativeEvent.composedPath().indexOf(event.currentTarget))
                .find(anyOf(isInstanceOf(HTMLAnchorElement), isInstanceOf(SVGAElement)))
            if (!anchor) {
                return
            }
            const href = typeof anchor.href === 'string' ? anchor.href : anchor.href.baseVal
            // Check if URL is outside the app
            if (!href.startsWith(window.location.origin)) {
                return
            }
            // Handle navigation programmatically
            event.preventDefault()
            history.push(new URL(href).pathname)
        },
        [history]
    )
    return (
        <RootComponent
            onClick={onClick}
            ref={refFn}
            className={classNames(className, 'markdown')}
            dangerouslySetInnerHTML={{ __html: dangerousInnerHTML }}
        />
    )
}
