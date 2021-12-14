import classNames from 'classnames'
import React from 'react'

import styles from '@sourcegraph/shared/src/components/Markdown.module.scss'

interface Props {
    wrapper?: 'div' | 'span'
    dangerousInnerHTML: string
    className?: string
    /** A function to attain a reference to the top-level div from a parent component. */
    refFn?: (reference: HTMLElement | null) => void
    testId?: string
}

export const Markdown: React.FunctionComponent<Props> = ({
    wrapper: RootComponent = 'div',
    refFn,
    className,
    dangerousInnerHTML,
    testId,
}: Props) => {
    // Remove links as they do not work in VS Code Web
    dangerousInnerHTML = dangerousInnerHTML.replace('href=', 'href="#" id="')
    // Links in markdown cannot use react-router's <Link>.
    // Prevent hitting the backend (full page reloads) for links that stay inside the app.

    return (
        <RootComponent
            data-testid={testId}
            // onClick={onClick}
            ref={refFn}
            className={classNames(className, styles.markdown)}
            dangerouslySetInnerHTML={{ __html: dangerousInnerHTML }}
        />
    )
}
