import { useCallback, useEffect, useMemo } from 'react'

import { isMacPlatform as isMacPlatformFunc } from '@sourcegraph/common'
import { isInputElement } from '@sourcegraph/shared/src/util/dom'

import { BlockDirection, BlockProps } from '..'

import { Notebook } from '.'

interface UseNotebookEventHandlersProps
    extends Pick<BlockProps, 'isReadOnly' | 'onMoveBlock' | 'onRunBlock' | 'onDeleteBlock' | 'onDuplicateBlock'> {
    notebook: Notebook
    selectedBlockId: string | null
    commandPaletteInputReference: React.RefObject<HTMLInputElement>
    selectBlock: (blockId: string | null) => void
}

export function focusBlockElement(blockId: string, isReadOnly: boolean): void {
    if (!isReadOnly) {
        document.querySelector<HTMLDivElement>(`[data-block-id="${blockId}"] .block`)?.focus()
    }
}

export function isModifierKeyPressed(isMetaKey: boolean, isCtrlKey: boolean, isMacPlatform: boolean): boolean {
    return (isMacPlatform && isMetaKey) || (!isMacPlatform && isCtrlKey)
}

export function useNotebookEventHandlers({
    notebook,
    selectedBlockId,
    commandPaletteInputReference,
    isReadOnly,
    selectBlock,
    onMoveBlock,
    onRunBlock,
    onDeleteBlock,
    onDuplicateBlock,
}: UseNotebookEventHandlersProps): void {
    const onMoveBlockSelection = useCallback(
        (id: string, direction: BlockDirection) => {
            const blockId = direction === 'up' ? notebook.getPreviousBlockId(id) : notebook.getNextBlockId(id)
            if (blockId) {
                selectBlock(blockId)
                focusBlockElement(blockId, isReadOnly)
            } else if (!blockId && direction === 'down') {
                commandPaletteInputReference.current?.focus()
            }
        },
        [notebook, commandPaletteInputReference, isReadOnly, selectBlock]
    )

    const isMacPlatform = useMemo(() => isMacPlatformFunc(), [])

    useEffect(() => {
        const handleMouseDownUpOrFocusIn = (event: MouseEvent | FocusEvent): void => {
            const target = event.target as HTMLElement | null
            const blockWrapper = target?.closest<HTMLDivElement>('.block-wrapper')
            if (!blockWrapper) {
                selectBlock(null)
                return
            }

            // When clicking buttons inside the block menu, wait for the mouseup
            // event to select the block to prevent buttons shifting.
            const blockMenu = target?.closest<HTMLDivElement>('.block-menu')
            if (blockMenu && event.type !== 'mouseup') {
                return
            }

            const blockId = blockWrapper.dataset.blockId
            if (blockId) {
                selectBlock(blockId)
            }
        }

        const handleKeyDown = (event: KeyboardEvent): void => {
            const target = event.target as HTMLElement

            if (isInputElement(target)) {
                return
            }

            if (!selectedBlockId && event.key === 'ArrowDown') {
                selectBlock(notebook.getFirstBlockId())
            } else if (event.key === 'Escape' && !isInputElement(target)) {
                selectBlock(null)
            }

            if (!selectedBlockId) {
                return
            }

            // Focus on the last `menuitem` of the prev block when using `Shift + Tab`
            // while focusing on selected block element
            if (
                document.activeElement ===
                    document.querySelector<HTMLDivElement>(`[data-block-id="${selectedBlockId}"] .block`) &&
                event.shiftKey &&
                event.key === 'Tab'
            ) {
                const previousBlockId = notebook.getPreviousBlockId(selectedBlockId)

                if (previousBlockId) {
                    event.preventDefault()

                    focusBlockElement(previousBlockId, isReadOnly)

                    const menuItems = document.querySelectorAll<HTMLAnchorElement>(
                        `[data-block-id="${previousBlockId}"] .block-menu [role="menuitem"]`
                    )
                    menuItems[menuItems.length - 1]?.focus()
                }
            }

            const isModifierKeyDown = isModifierKeyPressed(event.metaKey, event.ctrlKey, isMacPlatform)
            if (event.key === 'ArrowUp' || event.key === 'ArrowDown') {
                const direction = event.key === 'ArrowUp' ? 'up' : 'down'
                if (isModifierKeyDown) {
                    onMoveBlock(selectedBlockId, direction)
                    // Prevent page scrolling in Firefox
                    event.preventDefault()
                } else {
                    onMoveBlockSelection(selectedBlockId, direction)
                }
            } else if (event.key === 'Enter' && isModifierKeyDown) {
                onRunBlock(selectedBlockId)
            } else if (event.key === 'Delete' || (event.key === 'Backspace' && isModifierKeyDown)) {
                onDeleteBlock(selectedBlockId)
            } else if (event.key === 'd' && isModifierKeyDown) {
                event.preventDefault()
                onDuplicateBlock(selectedBlockId)
            }
        }

        document.addEventListener('keydown', handleKeyDown)
        // Check all clicks on the document and deselect the currently selected block if it was triggered outside of a block.
        document.addEventListener('mousedown', handleMouseDownUpOrFocusIn)
        document.addEventListener('mouseup', handleMouseDownUpOrFocusIn)
        // We're using the `focusin` event instead of the `focus` event, since the latter does not bubble up.
        document.addEventListener('focusin', handleMouseDownUpOrFocusIn)
        return () => {
            document.removeEventListener('keydown', handleKeyDown)
            document.removeEventListener('mousedown', handleMouseDownUpOrFocusIn)
            document.removeEventListener('mouseup', handleMouseDownUpOrFocusIn)
            document.removeEventListener('focusin', handleMouseDownUpOrFocusIn)
        }
    }, [
        notebook,
        selectedBlockId,
        isReadOnly,
        onMoveBlockSelection,
        selectBlock,
        isMacPlatform,
        onMoveBlock,
        onRunBlock,
        onDeleteBlock,
        onDuplicateBlock,
    ])
}
