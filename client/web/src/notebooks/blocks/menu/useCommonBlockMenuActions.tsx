import ArrowDownIcon from 'mdi-react/ArrowDownIcon'
import ArrowUpIcon from 'mdi-react/ArrowUpIcon'
import ContentDuplicateIcon from 'mdi-react/ContentDuplicateIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import React, { useMemo } from 'react'

import { Icon } from '@sourcegraph/wildcard'

import { BlockProps } from '../..'

import { BlockMenuAction } from './NotebookBlockMenu'

interface UseCommonBlockMenuActionsOptions
    extends Pick<BlockProps, 'isReadOnly' | 'onDeleteBlock' | 'onDuplicateBlock' | 'onMoveBlock'> {
    modifierKeyLabel: string
    isInputFocused: boolean
    isMacPlatform: boolean
}

export const useCommonBlockMenuActions = ({
    isInputFocused,
    isReadOnly,
    modifierKeyLabel,
    isMacPlatform,
    onMoveBlock,
    onDeleteBlock,
    onDuplicateBlock,
}: UseCommonBlockMenuActionsOptions): BlockMenuAction[] =>
    useMemo(() => {
        if (isReadOnly) {
            return []
        }
        return [
            {
                type: 'button',
                label: 'Duplicate',
                icon: <Icon as={ContentDuplicateIcon} />,
                onClick: onDuplicateBlock,
                keyboardShortcutLabel: !isInputFocused ? `${modifierKeyLabel} + D` : '',
            },
            {
                type: 'button',
                label: 'Move Up',
                icon: <Icon as={ArrowUpIcon} />,
                onClick: id => onMoveBlock(id, 'up'),
                keyboardShortcutLabel: !isInputFocused ? `${modifierKeyLabel} + ↑` : '',
            },
            {
                type: 'button',
                label: 'Move Down',
                icon: <Icon as={ArrowDownIcon} />,
                onClick: id => onMoveBlock(id, 'down'),
                keyboardShortcutLabel: !isInputFocused ? `${modifierKeyLabel} + ↓` : '',
            },
            {
                type: 'button',
                label: 'Delete',
                icon: <Icon as={DeleteIcon} />,
                onClick: onDeleteBlock,
                keyboardShortcutLabel: !isInputFocused ? (isMacPlatform ? '⌘ + ⌫' : 'Del') : '',
            },
        ]
    }, [isReadOnly, isMacPlatform, isInputFocused, modifierKeyLabel, onMoveBlock, onDeleteBlock, onDuplicateBlock])
