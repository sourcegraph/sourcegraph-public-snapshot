import { noop } from 'lodash'
import React, { useMemo } from 'react'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { NotebookBlock } from '@sourcegraph/shared/src/graphql/schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { SearchStreamingProps } from '..'
import { fetchRepository, resolveRevision } from '../../repo/backend'
import { StreamingSearchResultsListProps } from '../results/StreamingSearchResultsList'

import { SearchNotebook } from './SearchNotebook'

import { Block, BlockInit } from '.'

interface NotebookContentProps
    extends SearchStreamingProps,
        ThemeProps,
        TelemetryProps,
        Omit<StreamingSearchResultsListProps, 'allExpanded'>,
        ExtensionsControllerProps<'extHostAPI'> {
    globbing: boolean
    isMacPlatform: boolean
    viewerCanManage: boolean
    blocks: NotebookBlock[]
    onUpdateBlocks: (blocks: Block[]) => void
    fetchRepository: typeof fetchRepository
    resolveRevision: typeof resolveRevision
}

export const NotebookContent: React.FunctionComponent<NotebookContentProps> = ({
    viewerCanManage,
    blocks,
    onUpdateBlocks,
    resolveRevision,
    fetchRepository,
    ...props
}) => {
    const initializerBlocks: BlockInit[] = useMemo(
        () =>
            blocks.map(block => {
                switch (block.__typename) {
                    case 'MarkdownBlock':
                        return { id: block.id, type: 'md', input: block.markdownInput }
                    case 'QueryBlock':
                        return { id: block.id, type: 'query', input: block.queryInput }
                    case 'FileBlock':
                        return {
                            id: block.id,
                            type: 'file',
                            input: {
                                ...block.fileInput,
                                revision: block.fileInput.revision ?? '',
                            },
                        }
                }
            }),
        [blocks]
    )

    return (
        <SearchNotebook
            {...props}
            isReadOnly={!viewerCanManage}
            blocks={initializerBlocks}
            onSerializeBlocks={viewerCanManage ? onUpdateBlocks : noop}
            resolveRevision={resolveRevision}
            fetchRepository={fetchRepository}
        />
    )
}
