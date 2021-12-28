import { useCallback } from 'react'

import { BlockProps } from '.'

interface UseBlockShortcutsOptions
    extends Pick<
        BlockProps,
        'onMoveBlockSelection' | 'onDeleteBlock' | 'onRunBlock' | 'onDuplicateBlock' | 'onMoveBlock'
    > {
    id: string
    onEnterBlock: () => void
    isMacPlatform: boolean
}

export function isModifierKeyPressed(isMetaKey: boolean, isCtrlKey: boolean, isMacPlatform: boolean): boolean {
    return (isMacPlatform && isMetaKey) || (!isMacPlatform && isCtrlKey)
}

export const useBlockShortcuts = ({
    id,
    isMacPlatform,
    onMoveBlockSelection,
    onRunBlock,
    onDeleteBlock,
    onEnterBlock,
    onMoveBlock,
    onDuplicateBlock,
}: UseBlockShortcutsOptions): { onKeyDown: (event: React.KeyboardEvent) => void } => {
    const onKeyDown = useCallback(
        (event: React.KeyboardEvent): void => {
            const isModifierKeyDown = isModifierKeyPressed(event.metaKey, event.ctrlKey, isMacPlatform)
            if (event.key === 'ArrowUp' || event.key === 'ArrowDown') {
                const direction = event.key === 'ArrowUp' ? 'up' : 'down'
                if (isModifierKeyDown) {
                    onMoveBlock(id, direction)
                    // Prevent page scrolling in Firefox
                    event.preventDefault()
                } else {
                    onMoveBlockSelection(id, direction)
                }
            } else if (event.key === 'Enter') {
                if (isModifierKeyDown) {
                    onRunBlock(id)
                } else {
                    onEnterBlock()
                }
            } else if (event.key === 'Delete' || (event.key === 'Backspace' && isModifierKeyDown)) {
                onDeleteBlock(id)
            } else if (event.key === 'd' && isModifierKeyDown) {
                event.preventDefault()
                onDuplicateBlock(id)
            }
        },
        [
            id,
            isMacPlatform,
            onMoveBlockSelection,
            onRunBlock,
            onDeleteBlock,
            onEnterBlock,
            onMoveBlock,
            onDuplicateBlock,
        ]
    )

    return { onKeyDown }
}
