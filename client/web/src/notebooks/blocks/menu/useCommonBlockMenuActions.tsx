import React, { useMemo } from 'react'

import ArrowDownIcon from 'mdi-react/ArrowDownIcon'
import ArrowUpIcon from 'mdi-react/ArrowUpIcon'
import ContentDuplicateIcon from 'mdi-react/ContentDuplicateIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'

import { isMacPlatform as isMacPlatformFn } from '@sourcegraph/common'

import { BlockProps } from '../..'
import { useModifierKeyLabel } from '../useModifierKeyLabel'

import { BlockMenuAction } from './NotebookBlockMenu'

interface UseCommonBlockMenuActionsOptions
    extends Pick<BlockProps, 'isReadOnly' | 'onDeleteBlock' | 'onDuplicateBlock' | 'onMoveBlock'> {
    isInputFocused: boolean
}

export const useCommonBlockMenuActions = ({
    isInputFocused,
    isReadOnly,
    onMoveBlock,
    onDeleteBlock,
    onDuplicateBlock,
}: UseCommonBlockMenuActionsOptions): BlockMenuAction[] => {
    const isMacPlatform = useMemo(() => isMacPlatformFn(), [])
    const modifierKeyLabel = useModifierKeyLabel()
    return useMemo(() => {
        if (isReadOnly) {
            return []
        }
        return [
            {
                type: 'button',
                label: 'Duplicate',
                icon: <ContentDuplicateIcon className="icon-inline" />,
                onClick: onDuplicateBlock,
                keyboardShortcutLabel: !isInputFocused ? `${modifierKeyLabel} + D` : '',
            },
            {
                type: 'button',
                label: 'Move Up',
                icon: <ArrowUpIcon className="icon-inline" />,
                onClick: id => onMoveBlock(id, 'up'),
                keyboardShortcutLabel: !isInputFocused ? `${modifierKeyLabel} + ↑` : '',
            },
            {
                type: 'button',
                label: 'Move Down',
                icon: <ArrowDownIcon className="icon-inline" />,
                onClick: id => onMoveBlock(id, 'down'),
                keyboardShortcutLabel: !isInputFocused ? `${modifierKeyLabel} + ↓` : '',
            },
            {
                type: 'button',
                label: 'Delete',
                icon: <DeleteIcon className="icon-inline" />,
                onClick: onDeleteBlock,
                keyboardShortcutLabel: !isInputFocused ? (isMacPlatform ? '⌘ + ⌫' : 'Del') : '',
            },
        ]
    }, [isReadOnly, isMacPlatform, isInputFocused, modifierKeyLabel, onMoveBlock, onDeleteBlock, onDuplicateBlock])
}
