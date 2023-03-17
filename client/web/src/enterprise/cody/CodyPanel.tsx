import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CodyChat } from './CodyChat'

export const CodyPanel: React.FunctionComponent<
    {
        repoID: string
        repoName: string
        revision?: string
        filePath: string
        blobContent: string
    } & TelemetryProps
> = ({ repoName, filePath, blobContent, telemetryService }) => {
    useEffect(() => {
        telemetryService.log('CodyPanelOpened')
    }, [telemetryService])

    return (
        <div className="pt-3">
            <CodyChat
                promptPrefix={[
                    `Human: The content of the file '${filePath}' in repository '${repoName}' is`,
                    `<file>\n${blobContent}</file>`,
                    'Human: ',
                ].join('\n\n')}
            />
        </div>
    )
}
