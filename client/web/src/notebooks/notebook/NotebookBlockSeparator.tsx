import React from 'react'

import { BlockInput } from '..'

import { NotebookAddBlockButtons } from './NotebookAddBlockButtons'

import styles from './NotebookBlockSeparator.module.scss'

interface NotebookAddBlockButtonsProps {
    isReadOnly: boolean
    index: number
    onAddBlock: (blockIndex: number, blockInput: BlockInput) => void
}

export const NotebookBlockSeparator: React.FunctionComponent<
    React.PropsWithChildren<NotebookAddBlockButtonsProps>
> = React.memo(({ isReadOnly, index, onAddBlock }) =>
    isReadOnly ? (
        <div className="mb-2" />
    ) : (
        <div className={styles.blockSeparator}>
            <NotebookAddBlockButtons index={index} onAddBlock={onAddBlock} />
        </div>
    )
)
