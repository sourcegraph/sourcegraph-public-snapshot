import React, { useCallback, useEffect } from 'react'

import { BlockProps } from '.'

interface UseBlockFocusOptions extends Pick<BlockProps, 'onSelectBlock'> {
    id: string
    isSelected: boolean
    isInputFocused: boolean
    blockElement: HTMLElement | null
}

const isMonacoEditorDescendant = (element: HTMLElement): boolean => element.closest('.monaco-editor') !== null

export const useBlockSelection = ({
    id,
    isSelected,
    onSelectBlock,
    blockElement,
    isInputFocused,
}: UseBlockFocusOptions): {
    onBlur: (event: React.FocusEvent) => void
    onSelect: (event: React.MouseEvent | React.FocusEvent) => void
} => {
    const onBlur = useCallback(
        (event: React.FocusEvent) => {
            const relatedTarget = event.relatedTarget as HTMLElement
            if (!event.target.contains(relatedTarget)) {
                onSelectBlock(null)
            }
        },
        [onSelectBlock]
    )

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

    return { onBlur, onSelect }
}
