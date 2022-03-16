import React, { useMemo } from 'react'

import ArrowDownIcon from 'mdi-react/ArrowDownIcon'
import ArrowUpIcon from 'mdi-react/ArrowUpIcon'
import ContentDuplicateIcon from 'mdi-react/ContentDuplicateIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'

import { isMacPlatform as isMacPlatformFn } from '@sourcegraph/common'

import { BlockProps } from '../..'
import { useIsBlockInputFocused } from '../useIsBlockInputFocused'
import { useModifierKeyLabel } from '../useModifierKeyLabel'

import { BlockMenuAction } from './NotebookBlockMenu'

export const useCommonBlockMenuActions = ({
    id,
    isReadOnly,
    onMoveBlock,
    onDeleteBlock,
    onDuplicateBlock,
}: Pick<BlockProps, 'id' | 'isReadOnly' | 'onDeleteBlock' | 'onDuplicateBlock' | 'onMoveBlock'>): BlockMenuAction[] => {
    const isMacPlatform = useMemo(() => isMacPlatformFn(), [])
    const modifierKeyLabel = useModifierKeyLabel()
    const isInputFocused = useIsBlockInputFocused(id)
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
