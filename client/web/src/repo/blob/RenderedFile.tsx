import * as H from 'history'
import * as React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'

import { useScrollToLocationHash } from '../../components/useScrollToLocationHash'

import styles from './RenderedFile.module.scss'

interface Props {
    /**
     * The rendered HTML contents of the file.
     */
    dangerousInnerHTML: string

    location: H.Location
}

/**
 * Displays a file whose contents are rendered to HTML, such as a Markdown file.
 */
export const RenderedFile: React.FunctionComponent<Props> = ({ dangerousInnerHTML, location }) => {
    useScrollToLocationHash(location)
    return (
        <div className={styles.renderedFile}>
            <div className={styles.container}>
                <Markdown dangerousInnerHTML={dangerousInnerHTML} />
            </div>
        </div>
    )
}
