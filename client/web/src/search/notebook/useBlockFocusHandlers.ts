import React, { useCallback, useEffect } from 'react'

import { BlockProps } from '.'

type UseBlockFocusHandlersOptions = { isSelected: boolean; blockElement: HTMLElement | null } & Pick<
    BlockProps,
    'onSelectBlock'
>

export const useBlockFocusHandlers = ({
    isSelected,
    onSelectBlock,
    blockElement,
}: UseBlockFocusHandlersOptions): { onBlur: (event: React.FocusEvent) => void } => {
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
        if (isSelected) {
            blockElement?.focus()
        }
    }, [isSelected, blockElement])

    return { onBlur }
}
