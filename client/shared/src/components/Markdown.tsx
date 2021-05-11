import classNames from 'classnames'
import React, { useMemo } from 'react'
import { useHistory } from 'react-router'

import { createLinkClickHandler } from '../util/link-click-handler/linkClickHandler'

interface Props {
    wrapper?: 'div' | 'span'
    dangerousInnerHTML: string
    className?: string
    /** A function to attain a reference to the top-level div from a parent component. */
    refFn?: (reference: HTMLElement | null) => void
}

export const Markdown: React.FunctionComponent<Props> = ({
    wrapper: RootComponent = 'div',
    refFn,
    className,
    dangerousInnerHTML,
}: Props) => {
    const history = useHistory()

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
