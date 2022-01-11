import classNames from 'classnames'
import React from 'react'

import { Button } from '@sourcegraph/wildcard'

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
    <div
        className={classNames(styles.addBlockButtonsWrapper, !alwaysVisible && styles.showOnHover, className)}
        data-testid={alwaysVisible && 'always-visible-add-block-buttons'}
    >
        <hr className="mx-3" />
        <div className={styles.addBlockButtons}>
            <Button
                className={styles.addBlockButton}
                onClick={() => onAddBlock(index, { type: 'query', input: '// Enter search query' })}
                data-testid="add-query-button"
                outline={true}
                variant="secondary"
                size="sm"
            >
                + Query
            </Button>
            <Button
                className={classNames('ml-2', styles.addBlockButton)}
                onClick={() => onAddBlock(index, { type: 'md', input: '*Enter markdown*' })}
                data-testid="add-md-button"
                outline={true}
                variant="secondary"
                size="sm"
            >
                + Markdown
            </Button>
            <Button
                className={classNames('ml-2', styles.addBlockButton)}
                onClick={() =>
                    onAddBlock(index, {
                        type: 'file',
                        input: { repositoryName: '', revision: '', filePath: '', lineRange: null },
                    })
                }
                data-testid="add-file-button"
                outline={true}
                variant="secondary"
                size="sm"
            >
                + Code
            </Button>
        </div>
    </div>
)
