import React from 'react'
import { Markdown } from '../../../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../../../shared/src/util/markdown'
import { FileDiff } from '../../changes/computeDiff'

interface Props {
    // TODO!(sqs): cant show file create/rename/delete operations unless we use our internal
    // WorkspaceEdit type's #operations field.
    fileDiffs: FileDiff[]

    className?: string
}

/**
 * Previews a workspace edit's changes.
 */
export const WorkspaceEditPreview: React.FunctionComponent<Props> = ({ fileDiffs, className = '' }) => {
    const rawDiff = fileDiffs.map(d => d.hunks.map(h => h.body).join('\n')).join('\n\n')
    return <Markdown dangerousInnerHTML={renderMarkdown('```diff\n' + rawDiff + '\n```')} className={className} />
}
