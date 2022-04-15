import React, { useEffect } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container } from '@sourcegraph/wildcard'

import { NotebooksOverview } from './NotebooksOverview'

interface NotebooksGettingStartedTabProps extends TelemetryProps {}

const panels = [
    {
        title: 'Enable the notepad',
        description:
            "Create notebooks while you browse. Add searches, files, and file ranges without leaving the page you're on, then create a notebook with one click.",
        videoSources: [
            { type: 'webm', src: 'https://storage.googleapis.com/sourcegraph-assets/batch-changes/how-it-works.webm' },
            { type: 'mp4', src: 'https://storage.googleapis.com/sourcegraph-assets/batch-changes/how-it-works.mp4' },
        ],
    },
    {
        title: 'The command palette',
        description:
            'Use slash commands to choose from the available block Notebook block types. Markdown, file, symbol, and search query blocks are supported.',
        videoSources: [
            {
                type: 'webm',
                src: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_command_palette.webm',
            },
            {
                type: 'mp4',
                src: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_command_palette.mp4',
            },
        ],
    },
    {
        title: 'Keep your docs current-automatically-with symbol blocks',
        description:
            'Symbol blocks follow a chosen symbol anywhere in a file, even as it changes. Create symbol blocks to keep your docs from getting stale.',
        videoSources: [
            {
                type: 'webm',
                src: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_symbol_block.webm',
            },
            {
                type: 'mp4',
                src: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_symbol_block.mp4',
            },
        ],
    },
    {
        title: 'Share Notebooks with your team or company',
        description:
            "Notebooks are private by default, but you can share them with your team (if you're using Sourcegraph organizations) or with your company.",
        videoSources: [
            { type: 'webm', src: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_sharing.webm' },
            { type: 'mp4', src: 'https://storage.googleapis.com/sourcegraph-assets/notebooks/notebooks_sharing.mp4' },
        ],
    },
]

export const NotebooksGettingStartedTab: React.FunctionComponent<NotebooksGettingStartedTabProps> = ({
    telemetryService,
}) => {
    useEffect(() => telemetryService.log('NotebooksGettingStartedTabViewed'), [telemetryService])

    return (
        <>
            <NotebooksOverview />
            <h3>Functionality</h3>
            <div className="row">
                {panels.map(panel => (
                    <div key={panel.title} className="col-12 col-md-6 p-2">
                        <Container>
                            <video
                                className="w-100 h-auto shadow percy-hide"
                                autoPlay={false}
                                muted={true}
                                loop={false}
                                playsInline={true}
                                controls={true}
                            >
                                {panel.videoSources.map(videoSource => (
                                    <source
                                        key={videoSource.src}
                                        type={`video/${videoSource.type}`}
                                        src={videoSource.src}
                                    />
                                ))}
                            </video>
                            <div className="mt-3 mb-2">
                                <strong>{panel.title}</strong>
                            </div>
                            <p>{panel.description}</p>
                        </Container>
                    </div>
                ))}
            </div>
        </>
    )
}
