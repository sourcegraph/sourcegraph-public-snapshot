import { forwardRef } from 'react'

import classNames from 'classnames'
import * as H from 'history'

import { Markdown } from '@sourcegraph/wildcard'

import { useScrollToLocationHash } from '../../components/useScrollToLocationHash'

import styles from './RenderedFile.module.scss'

interface Props {
    /** The rendered HTML contents of the file. */
    dangerousInnerHTML: string
    location: H.Location
    className?: string
}

/**
 * Displays a file whose contents are rendered to HTML, such as a Markdown file.
 */
export const RenderedFile = forwardRef<HTMLDivElement, Props>((props, reference) => {
    const { dangerousInnerHTML, location, className } = props

    useScrollToLocationHash(location)

    return (
        <div ref={reference} className={classNames(styles.renderedFile, className)}>
            <div className={styles.container}>
                <Markdown dangerousInnerHTML={dangerousInnerHTML} />
            </div>
        </div>
    )
})
