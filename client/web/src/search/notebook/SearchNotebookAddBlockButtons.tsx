import classNames from 'classnames'
import React from 'react'

import styles from './SearchNotebookAddBlockButtons.module.scss'

import { BlockType } from '.'

interface SearchNotebookAddBlockButtonsProps {
    onAddBlock: (blockIndex: number, type: BlockType, input: string) => void
    index: number
    alwaysVisible?: boolean
    className?: string
}

export const SearchNotebookAddBlockButtons: React.FunctionComponent<SearchNotebookAddBlockButtonsProps> = ({
    alwaysVisible,
    index,
    className,
    onAddBlock,
}) => (
    <div className={classNames(styles.addBlockButtonsWrapper, !alwaysVisible && 'show-on-hover', className)}>
        <hr className="mx-3" />
        <div className={styles.addBlockButtons}>
            <button
                type="button"
                className="btn btn-secondary btn-sm mr-2"
                onClick={() => onAddBlock(index, 'query', '// Enter search query')}
            >
                + Query
            </button>
            <button
                type="button"
                className="btn btn-sm btn-secondary"
                onClick={() => onAddBlock(index, 'md', '*Enter markdown*')}
            >
                + Markdown
            </button>
        </div>
    </div>
)
