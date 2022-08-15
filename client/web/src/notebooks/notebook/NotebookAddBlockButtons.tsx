import React, { useCallback } from 'react'

import { mdiLanguageMarkdownOutline, mdiMagnify, mdiCodeTags, mdiFunction, mdiLaptop } from '@mdi/js'

import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

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
            <Tooltip content="Add Markdown text">
                <Button
                    className={styles.addBlockButton}
                    onClick={() => addBlock({ type: 'md', input: { text: '', initialFocusInput: true } })}
                    data-testid="add-md-block"
                    aria-label="Add markdown"
                >
                    <Icon aria-hidden={true} size="sm" svgPath={mdiLanguageMarkdownOutline} />
                </Button>
            </Tooltip>
            <Tooltip content="Add a Sourcegraph query">
                <Button
                    className={styles.addBlockButton}
                    onClick={() => addBlock({ type: 'query', input: { query: '', initialFocusInput: true } })}
                    data-testid="add-query-block"
                    aria-label="Add query"
                >
                    <Icon aria-hidden={true} size="sm" svgPath={mdiMagnify} />
                </Button>
            </Tooltip>
            <Tooltip content="Add code from a file">
                <Button
                    className={styles.addBlockButton}
                    onClick={() => addBlock({ type: 'file', input: EMPTY_FILE_BLOCK_INPUT })}
                    data-testid="add-file-block"
                    aria-label="Add code"
                >
                    <Icon aria-hidden={true} size="sm" svgPath={mdiCodeTags} />
                </Button>
            </Tooltip>
            <Tooltip content="Add a symbol">
                <Button
                    className={styles.addBlockButton}
                    onClick={() => addBlock({ type: 'symbol', input: EMPTY_SYMBOL_BLOCK_INPUT })}
                    data-testid="add-symbol-block"
                    aria-label="Add symbol"
                >
                    <Icon aria-hidden={true} size="sm" svgPath={mdiFunction} />
                </Button>
            </Tooltip>
            {showComputeComponent && (
                <Tooltip content="Add compute block">
                    <Button
                        className={styles.addBlockButton}
                        onClick={() => addBlock({ type: 'compute', input: '' })}
                        data-testid="add-compute-block"
                        aria-label="Add compute block"
                    >
                        {/* // TODO: Fix icon */}
                        <Icon aria-hidden={true} size="sm" svgPath={mdiLaptop} />
                    </Button>
                </Tooltip>
            )}
        </>
    )
}
