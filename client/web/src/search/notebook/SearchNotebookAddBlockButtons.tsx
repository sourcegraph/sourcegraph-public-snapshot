import classNames from 'classnames'
import React from 'react'

import styles from './SearchNotebookAddBlockButtons.module.scss'

import { BlockInput } from '.'

interface SearchNotebookAddBlockButtonsProps {
    onAddBlock: (blockIndex: number, blockInput: BlockInput) => void
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
    <div className={classNames(styles.addBlockButtonsWrapper, !alwaysVisible && styles.showOnHover, className)}>
        <hr className="mx-3" />
        <div className={styles.addBlockButtons}>
            <button
                type="button"
                className={classNames('btn btn-outline-secondary btn-sm mr-2', styles.addBlockButton)}
                onClick={() => onAddBlock(index, { type: 'query', input: '// Enter search query' })}
                data-testid="add-query-button"
            >
                + Query
            </button>
            <button
                type="button"
                className={classNames('btn btn-outline-secondary btn-sm', styles.addBlockButton)}
                onClick={() => onAddBlock(index, { type: 'md', input: '*Enter markdown*' })}
                data-testid="add-md-button"
            >
                + Markdown
            </button>
        </div>
    </div>
)
