import { useCallback, useEffect, useMemo } from 'react'

import { isMacPlatform as isMacPlatformFn } from '@sourcegraph/common'
import { isInputElement } from '@sourcegraph/shared/src/util/dom'

import { BlockDirection, BlockProps } from '..'

import { Notebook } from '.'

interface UseNotebookEventHandlersProps
    extends Pick<BlockProps, 'onMoveBlock' | 'onRunBlock' | 'onDeleteBlock' | 'onDuplicateBlock'> {
    notebook: Notebook
    selectedBlockId: string | null
    commandPaletteInputReference: React.RefObject<HTMLInputElement>
    setSelectedBlockId: (blockId: string | null) => void
}

export function focusBlock(blockId: string): void {
    document.querySelector<HTMLDivElement>(`[data-block-id="${blockId}"] .block`)?.focus()
}

export function isModifierKeyPressed(isMetaKey: boolean, isCtrlKey: boolean, isMacPlatform: boolean): boolean {
    return (isMacPlatform && isMetaKey) || (!isMacPlatform && isCtrlKey)
}

export const isMonacoEditorDescendant = (element: HTMLElement): boolean => element.closest('.monaco-editor') !== null

export function useNotebookEventHandlers({
    notebook,
    selectedBlockId,
    commandPaletteInputReference,
    setSelectedBlockId,
    onMoveBlock,
    onRunBlock,
    onDeleteBlock,
    onDuplicateBlock,
}: UseNotebookEventHandlersProps): void {
    const onMoveBlockSelection = useCallback(
        (id: string, direction: BlockDirection) => {
            const blockId = direction === 'up' ? notebook.getPreviousBlockId(id) : notebook.getNextBlockId(id)
            if (blockId) {
                setSelectedBlockId(blockId)
                focusBlock(blockId)
            } else if (!blockId && direction === 'down') {
                commandPaletteInputReference.current?.focus()
            }
        },
        [notebook, commandPaletteInputReference, setSelectedBlockId]
    )

    const isMacPlatform = useMemo(() => isMacPlatformFn(), [])

    useEffect(() => {
        const handleMouseDownUpOrFocusIn = (event: MouseEvent | FocusEvent): void => {
            const target = event.target as HTMLElement | null
            const blockWrapper = target?.closest<HTMLDivElement>('.block-wrapper')
            if (!blockWrapper) {
                setSelectedBlockId(null)
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
                setSelectedBlockId(blockId)
            }
        }

        const handleKeyDown = (event: KeyboardEvent): void => {
            const target = event.target as HTMLElement

            if (isInputElement(target)) {
                return
            }

            if (!selectedBlockId && event.key === 'ArrowDown') {
                setSelectedBlockId(notebook.getFirstBlockId())
            } else if (
                event.key === 'Escape' &&
                !isMonacoEditorDescendant(target) &&
                target.tagName.toLowerCase() !== 'input'
            ) {
                setSelectedBlockId(null)
            }

            if (!selectedBlockId) {
                return
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
        onMoveBlockSelection,
        setSelectedBlockId,
        isMacPlatform,
        onMoveBlock,
        onRunBlock,
        onDeleteBlock,
        onDuplicateBlock,
    ])
}
