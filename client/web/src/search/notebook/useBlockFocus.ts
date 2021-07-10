import React, { useCallback, useEffect } from 'react'

import { BlockProps } from '.'

interface UseBlockFocusOptions extends Pick<BlockProps, 'onSelectBlock'> {
    isSelected: boolean
    isInputFocused: boolean
    blockElement: HTMLElement | null
}

export const useBlockFocus = ({
    isSelected,
    onSelectBlock,
    blockElement,
    isInputFocused,
}: UseBlockFocusOptions): { onBlur: (event: React.FocusEvent) => void } => {
    const onBlur = useCallback(
        (event: React.FocusEvent) => {
            const relatedTarget = event.relatedTarget as HTMLElement
            if (!event.target.contains(relatedTarget)) {
                onSelectBlock(null)
            }
        },
        [onSelectBlock]
    )

    useEffect(() => {
        if (isSelected && !isInputFocused) {
            blockElement?.focus()
        }
    }, [isSelected, blockElement, isInputFocused])

    return { onBlur }
}
