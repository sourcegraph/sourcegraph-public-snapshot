import React from 'react'

import { mdiFileDocumentOutline, mdiMagnify } from '@mdi/js'

import { ContextFile } from '@sourcegraph/cody-shared/src/codebase-context/messages'
import { pluralize } from '@sourcegraph/common'

import { TranscriptAction } from './actions/TranscriptAction'

export interface FileLinkProps {
    path: string
    repoName?: string
    revision?: string
}

export const ContextFiles: React.FunctionComponent<{
    contextFiles: ContextFile[]
    fileLinkComponent: React.FunctionComponent<FileLinkProps>
    className?: string
}> = ({ contextFiles, fileLinkComponent: FileLink, className }) => (
    <TranscriptAction
        title={{ verb: 'Read', object: `${contextFiles.length} ${pluralize('file', contextFiles.length)}` }}
        steps={[
            { verb: 'Searched', object: 'entire codebase for relevant files', icon: mdiMagnify },
            ...contextFiles.map(file => ({
                verb: '',
                object: <FileLink path={file.fileName} repoName={file.repoName} revision={file.revision} />,
                icon: mdiFileDocumentOutline,
            })),
        ]}
        className={className}
    />
)
