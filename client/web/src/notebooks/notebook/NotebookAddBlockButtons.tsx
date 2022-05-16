import React, { useCallback } from 'react'

import CodeTagsIcon from 'mdi-react/CodeTagsIcon'
import FunctionIcon from 'mdi-react/FunctionIcon'
import LanguageMarkdownOutlineIcon from 'mdi-react/LanguageMarkdownOutlineIcon'
import LaptopIcon from 'mdi-react/LaptopIcon'
import MagnifyIcon from 'mdi-react/MagnifyIcon'

import { Button, Icon } from '@sourcegraph/wildcard'

import { BlockInput } from '..'
import { useExperimentalFeatures } from '../../stores'

import { EMPTY_FILE_BLOCK_INPUT, EMPTY_SYMBOL_BLOCK_INPUT } from './useCommandPaletteOptions'

import styles from './NotebookAddBlockButtons.module.scss'

interface NotebookAddBlockButtonsProps {
    onAddBlock: (blockIndex: number, blockInput: BlockInput) => void
    index: number
}

export const NotebookAddBlockButtons: React.FunctionComponent<
    React.PropsWithChildren<NotebookAddBlockButtonsProps>
> = ({ index, onAddBlock }) => {
    const showComputeComponent = useExperimentalFeatures(features => features.showComputeComponent)
    const addBlock = useCallback((blockInput: BlockInput) => onAddBlock(index, blockInput), [index, onAddBlock])
    return (
        <>
            <Button
                className={styles.addBlockButton}
                data-tooltip="Add Markdown text"
                aria-label="Add markdown"
                onClick={() => addBlock({ type: 'md', input: { text: '', initialFocusInput: true } })}
                data-testid="add-md-block"
            >
                <Icon as={LanguageMarkdownOutlineIcon} size="sm" />
            </Button>
            <Button
                className={styles.addBlockButton}
                data-tooltip="Add a Sourcegraph query"
                aria-label="Add query"
                onClick={() => addBlock({ type: 'query', input: { query: '', initialFocusInput: true } })}
                data-testid="add-query-block"
            >
                <Icon as={MagnifyIcon} size="sm" />
            </Button>
            <Button
                className={styles.addBlockButton}
                data-tooltip="Add code from a file"
                aria-label="Add code from file"
                onClick={() => addBlock({ type: 'file', input: EMPTY_FILE_BLOCK_INPUT })}
                data-testid="add-file-block"
            >
                <Icon as={CodeTagsIcon} size="sm" />
            </Button>
            <Button
                className={styles.addBlockButton}
                data-tooltip="Add a symbol"
                aria-label="Add symbol"
                onClick={() => addBlock({ type: 'symbol', input: EMPTY_SYMBOL_BLOCK_INPUT })}
                data-testid="add-symbol-block"
            >
                <Icon as={FunctionIcon} size="sm" />
            </Button>
            {showComputeComponent && (
                <Button
                    className={styles.addBlockButton}
                    data-tooltip="Add compute block"
                    aria-label="Add compute block"
                    onClick={() => addBlock({ type: 'compute', input: '' })}
                    data-testid="add-compute-block"
                >
                    {/* // TODO: Fix icon */}
                    <Icon as={LaptopIcon} size="sm" />
                </Button>
            )}
        </>
    )
}
