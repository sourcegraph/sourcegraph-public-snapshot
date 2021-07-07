import React, { useCallback, useMemo, useState } from 'react'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { SearchStreamingProps } from '..'
import { StreamingSearchResultsListProps } from '../results/StreamingSearchResultsList'

import { SearchNotebookAddBlockButtons } from './SearchNotebookAddBlockButtons'
import { SearchNotebookMarkdownBlock } from './SearchNotebookMarkdownBlock'
import { SearchNotebookQueryBlock } from './SearchNotebookQueryBlock'

import { Block, BlockType, Notebook } from '.'

interface SearchNotebookProps
    extends SearchStreamingProps,
        ThemeProps,
        Omit<StreamingSearchResultsListProps, 'allExpanded'> {
    globbing: boolean

    onBlocksChange: (blocks: Block[]) => void
    blocks: Omit<Block, 'id' | 'output'>[]
}

export const SearchNotebook: React.FunctionComponent<SearchNotebookProps> = ({ onBlocksChange, ...props }) => {
    const notebook = useMemo(() => new Notebook(props.blocks), [props.blocks])

    const [blocks, setBlocks] = useState<Block[]>(notebook.getBlocks())

    const onRunBlock = useCallback(
        (id: string) => {
            notebook.runBlockById(id)
            const blocks = notebook.getBlocks()
            setBlocks(blocks)
            onBlocksChange(blocks)
        },
        [notebook, onBlocksChange]
    )

    const onBlockInputChange = useCallback(
        (id: string, value: string) => {
            notebook.setBlockInputById(id, value)
            setBlocks(notebook.getBlocks())
        },
        [notebook]
    )

    const addBlock = useCallback(
        (index: number, type: BlockType, input: string) => {
            const addedBlock = notebook.insertBlockAtIndex(index, type, input)
            if (addedBlock.type === 'md') {
                notebook.runBlockById(addedBlock.id)
            }
            setBlocks(notebook.getBlocks())
        },
        [notebook, setBlocks]
    )

    return (
        <div className="w-100">
            {blocks.map((block, blockIndex) => (
                <div key={block.id}>
                    <SearchNotebookAddBlockButtons onAddBlock={addBlock} index={blockIndex} />
                    <>
                        {block.type === 'md' && (
                            <SearchNotebookMarkdownBlock
                                {...props}
                                {...block}
                                onRunBlock={onRunBlock}
                                onBlockInputChange={onBlockInputChange}
                            />
                        )}
                        {block.type === 'query' && (
                            <SearchNotebookQueryBlock
                                {...props}
                                {...block}
                                onRunBlock={onRunBlock}
                                onBlockInputChange={onBlockInputChange}
                            />
                        )}
                    </>
                </div>
            ))}
            <SearchNotebookAddBlockButtons
                onAddBlock={addBlock}
                index={blocks.length}
                className="mt-4"
                alwaysVisible={true}
            />
        </div>
    )
}
