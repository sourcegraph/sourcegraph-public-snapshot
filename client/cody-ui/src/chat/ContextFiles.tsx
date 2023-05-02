import React from 'react'

import { mdiFileDocumentOutline, mdiMagnify } from '@mdi/js'

import { pluralize } from '@sourcegraph/common'

import { TranscriptAction } from './actions/TranscriptAction'

export interface FileLinkProps {
    path: string
}

export const ContextFiles: React.FunctionComponent<{
    contextFiles: string[]
    fileLinkComponent: React.FunctionComponent<FileLinkProps>
    className?: string
}> = ({ contextFiles, fileLinkComponent: FileLink, className }) => (
    <TranscriptAction
        title={{ verb: 'Read', object: `${contextFiles.length} ${pluralize('file', contextFiles.length)}` }}
        steps={[
            { verb: 'Searched', object: 'entire codebase for relevant files', icon: mdiMagnify },
            ...contextFiles.map(file => ({
                verb: 'Read',
                object: <FileLink path={file} />,
                icon: mdiFileDocumentOutline,
            })),
        ]}
        className={className}
    />
)
