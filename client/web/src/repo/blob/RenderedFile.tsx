import * as H from 'history'
import * as React from 'react'

import { Markdown } from '@sourcegraph/shared/src/components/Markdown'

import { useScrollToLocationHash } from '../../components/useScrollToLocationHash'

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
        <div className="rendered-file">
            <div className="rendered-file__container">
                <Markdown dangerousInnerHTML={dangerousInnerHTML} />
            </div>
        </div>
    )
}
