import React, { useMemo } from 'react'

import classNames from 'classnames'
import { useHistory } from 'react-router'

import { createLinkClickHandler } from './utils/linkClickHandler'

import styles from './Markdown.module.scss'

interface Props {
    wrapper?: 'div' | 'span'
    dangerousInnerHTML: string
    className?: string
    testId?: string
}

export const Markdown = React.forwardRef<HTMLElement, Props>(
    ({ wrapper: RootComponent = 'div', className, dangerousInnerHTML, testId }, reference) => {
        const history = useHistory()

        // Links in markdown cannot use react-router's <Link>.
        // Prevent hitting the backend (full page reloads) for links that stay inside the app.
        const onClick = useMemo(() => createLinkClickHandler(history), [history])
        return (
            <RootComponent
                data-testid={testId}
                onClick={onClick}
                className={classNames(className, styles.markdown)}
                dangerouslySetInnerHTML={{ __html: dangerousInnerHTML }}
                // This casting is necessary due to TypeScript not being able
                // to understand React.forwardRef with generic elements.
                ref={reference as React.Ref<HTMLDivElement>}
            />
        )
    }
)
