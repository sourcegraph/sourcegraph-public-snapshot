import React, { useMemo } from 'react'
import classNames from 'classnames'
import { History } from 'history'
import { createLinkClickHandler } from 'utils/linkClickHandler'

interface Props {
    wrapper?: 'div' | 'span'
    dangerousInnerHTML: string
    history: History
    className?: string
    /** A function to attain a reference to the top-level div from a parent component. */
    refFn?: (reference: HTMLElement | null) => void
}

export const Markdown: React.FC<Props> = props => {
    const { wrapper: RootComponent = 'div', refFn, className, history, dangerousInnerHTML } = props

    // Links in markdown cannot use react-router's <Link>.
    // Prevent hitting the backend (full page reloads) for links that stay inside the app.
    const onClick = useMemo(() => createLinkClickHandler(history), [history])

    return (
        <RootComponent
            onClick={onClick}
            ref={refFn}
            className={classNames(className, 'markdown')}
            dangerouslySetInnerHTML={{ __html: dangerousInnerHTML }}
        />
    )
}
