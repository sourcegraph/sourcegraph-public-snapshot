import React, { useCallback, useEffect } from 'react'

import { Block, BlockProps } from '..'

interface UseBlockFocusOptions extends Pick<BlockProps<Block>, 'onSelectBlock'> {
    id: string
    isSelected: boolean
    isInputFocused: boolean
    blockElement: HTMLElement | null
}

export const isMonacoEditorDescendant = (element: HTMLElement): boolean => element.closest('.monaco-editor') !== null

export const useBlockSelection = ({
    id,
    isSelected,
    onSelectBlock,
    blockElement,
    isInputFocused,
}: UseBlockFocusOptions): {
    onSelect: (event: React.MouseEvent | React.FocusEvent) => void
} => {
    const onSelect = useCallback(
        (event: React.MouseEvent | React.FocusEvent) => {
            // Let Monaco input handle focus/click events
            if (isMonacoEditorDescendant(event.target as HTMLElement)) {
                return
            }
            onSelectBlock(id)
        },
        [id, onSelectBlock]
    )

    useEffect(() => {
        if (isSelected && !isInputFocused) {
            blockElement?.focus()
        } else if (!isSelected) {
            blockElement?.blur()
        }
    }, [isSelected, blockElement, isInputFocused])

    return { onSelect }
}
