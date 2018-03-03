import * as React from 'react'
import { Markdown } from '../../components/Markdown'

interface Props {
    /**
     * The rendered HTML contents of the file.
     */
    dangerousInnerHTML: string
}

/**
 * Displays a file whose contents are rendered to HTML, such as a Markdown file.
 */
export const RenderedFile: React.SFC<Props> = ({ dangerousInnerHTML }) => (
    <div className="rendered-file">
        <div className="rendered-file__container">
            <Markdown dangerousInnerHTML={dangerousInnerHTML} />
        </div>
    </div>
)
