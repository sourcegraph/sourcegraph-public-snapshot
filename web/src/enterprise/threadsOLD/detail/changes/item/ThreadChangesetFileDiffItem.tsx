import FileIcon from 'mdi-react/FileIcon'
import React from 'react'
import { Markdown } from '../../../../../../../shared/src/components/Markdown'
import { displayRepoName } from '../../../../../../../shared/src/components/RepoFileLink'
import { renderMarkdown } from '../../../../../../../shared/src/util/markdown'
import { parseRepoURI } from '../../../../../../../shared/src/util/url'
import { FileDiff } from '../computeDiff'

interface Props {
    fileDiff: FileDiff
    className?: string
    headerClassName?: string
    headerStyle?: React.CSSProperties
}

/**
 * A file diff in a thread changeset.
 */
export const ThreadChangesetFileDiffItem: React.FunctionComponent<Props> = ({
    fileDiff,
    className = '',
    headerClassName = '',
    headerStyle,
}) => {
    const parsed = parseRepoURI(fileDiff.newPath || fileDiff.oldPath!)
    return (
        <div className={`card border ${className}`}>
            <header className={`card-header d-flex align-items-start ${headerClassName}`} style={headerStyle}>
                <div className="flex-1 d-flex align-items-center">
                    <FileIcon className="icon-inline mr-2" />
                    <h3 className="mb-0 h6">{parsed.filePath}</h3>
                </div>
            </header>
            <Markdown
                dangerousInnerHTML={renderMarkdown('```diff\n' + fileDiff.hunks.map(h => h.body).join('\n') + '\n```')}
                className="overflow-auto"
            />
        </div>
    )
}
