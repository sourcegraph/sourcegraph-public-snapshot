import React, { useMemo } from 'react'

import { noop } from 'lodash'
import { Observable } from 'rxjs'

import { StreamingSearchResultsListProps } from '@sourcegraph/branded'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { Block, BlockInit } from '..'
import { NotebookFields } from '../../graphql-operations'
import { SearchStreamingProps } from '../../search'
import { CopyNotebookProps } from '../notebook'
import { NotebookComponent } from '../notebook/NotebookComponent'

export interface NotebookContentProps
    extends SearchStreamingProps,
        ThemeProps,
        TelemetryProps,
        Omit<StreamingSearchResultsListProps, 'allExpanded' | 'platformContext' | 'executedQuery'>,
        PlatformContextProps<'sourcegraphURL' | 'requestGraphQL' | 'urlToFile' | 'settings'> {
    authenticatedUser: AuthenticatedUser | null
    globbing: boolean
    viewerCanManage: boolean
    blocks: NotebookFields['blocks']
    exportedFileName: string
    isEmbedded?: boolean
    outlineContainerElement?: HTMLElement | null
    onUpdateBlocks: (blocks: Block[]) => void
    onCopyNotebook: (props: Omit<CopyNotebookProps, 'title'>) => Observable<NotebookFields>
}

export const NotebookContent: React.FunctionComponent<React.PropsWithChildren<NotebookContentProps>> = React.memo(
    ({
        viewerCanManage,
        blocks,
        exportedFileName,
        onCopyNotebook,
        onUpdateBlocks,
        globbing,
        streamSearch,
        isLightTheme,
        telemetryService,
        searchContextsEnabled,
        isSourcegraphDotCom,
        fetchHighlightedFileLineRanges,
        authenticatedUser,
        settingsCascade,
        platformContext,
        outlineContainerElement,
        isEmbedded,
    }) => {
        const initializerBlocks: BlockInit[] = useMemo(
            () =>
                blocks.map(block => {
                    switch (block.__typename) {
                        case 'MarkdownBlock':
                            return { id: block.id, type: 'md', input: { text: block.markdownInput } }
                        case 'QueryBlock':
                            return { id: block.id, type: 'query', input: { query: block.queryInput } }
                        case 'FileBlock':
                            return {
                                id: block.id,
                                type: 'file',
                                input: { ...block.fileInput, revision: block.fileInput.revision ?? '' },
                            }
                        case 'SymbolBlock':
                            return {
                                id: block.id,
                                type: 'symbol',
                                input: { ...block.symbolInput, revision: block.symbolInput.revision ?? '' },
                            }
                    }
                }),
            [blocks]
        )

        return (
            <NotebookComponent
                globbing={globbing}
                streamSearch={streamSearch}
                isLightTheme={isLightTheme}
                telemetryService={telemetryService}
                searchContextsEnabled={searchContextsEnabled}
                isSourcegraphDotCom={isSourcegraphDotCom}
                fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                authenticatedUser={authenticatedUser}
                settingsCascade={settingsCascade}
                platformContext={platformContext}
                isReadOnly={!viewerCanManage}
                blocks={initializerBlocks}
                onSerializeBlocks={viewerCanManage ? onUpdateBlocks : noop}
                exportedFileName={exportedFileName}
                onCopyNotebook={onCopyNotebook}
                outlineContainerElement={outlineContainerElement}
                isEmbedded={isEmbedded}
            />
        )
    }
)

NotebookContent.displayName = 'NotebookContent'
