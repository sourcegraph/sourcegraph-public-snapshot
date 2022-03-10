import classNames from 'classnames'
import React from 'react'

import { Button } from '@sourcegraph/wildcard'

import { BlockInput } from '..'
import { useExperimentalFeatures } from '../../stores'

import styles from './NotebookAddBlockButtons.module.scss'

interface NotebookAddBlockButtonsProps {
    onAddBlock: (blockIndex: number, blockInput: BlockInput) => void
    index: number
    alwaysVisible?: boolean
    className?: string
}

export const NotebookAddBlockButtons: React.FunctionComponent<NotebookAddBlockButtonsProps> = ({
    alwaysVisible,
    index,
    className,
    onAddBlock,
}) => {
    const showComputeComponent = useExperimentalFeatures(features => features.showComputeComponent)
    return (
        <div
            className={classNames(styles.addBlockButtonsWrapper, !alwaysVisible && styles.showOnHover, className)}
            data-testid={alwaysVisible && 'always-visible-add-block-buttons'}
        >
            <hr className="mx-3" />
            <div className={styles.addBlockButtons}>
                <Button
                    className={styles.addBlockButton}
                    onClick={() => onAddBlock(index, { type: 'query', input: '' })}
                    data-testid="add-query-button"
                    outline={true}
                    variant="secondary"
                    size="sm"
                >
                    + Query
                </Button>
                <Button
                    className={classNames('ml-2', styles.addBlockButton)}
                    onClick={() => onAddBlock(index, { type: 'md', input: '' })}
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
                {showComputeComponent ? (
                    <Button
                        className={classNames('ml-2', styles.addBlockButton)}
                        onClick={() =>
                            onAddBlock(index, {
                                type: 'compute',
                                input: 'placeholder',
                            })
                        }
                        data-testid="add-compute-button"
                        outline={true}
                        variant="secondary"
                        size="sm"
                    >
                        + Compute
                    </Button>
                ) : (
                    <div />
                )}
            </div>
        </div>
    )
}
