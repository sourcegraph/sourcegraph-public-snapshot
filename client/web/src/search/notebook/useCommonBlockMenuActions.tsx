import ArrowDownIcon from 'mdi-react/ArrowDownIcon'
import ArrowUpIcon from 'mdi-react/ArrowUpIcon'
import ContentDuplicateIcon from 'mdi-react/ContentDuplicateIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import React, { useMemo } from 'react'

import { BlockMenuAction } from './SearchNotebookBlockMenu'

import { BlockProps } from '.'

interface UseCommonBlockMenuActionsOptions
    extends Pick<BlockProps, 'isReadOnly' | 'onDeleteBlock' | 'onDuplicateBlock' | 'onMoveBlock'> {
    modifierKeyLabel: string
    isInputFocused: boolean
}

export const useCommonBlockMenuActions = ({
    isInputFocused,
    isReadOnly,
    modifierKeyLabel,
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
                label: 'Duplicate',
                icon: <ContentDuplicateIcon className="icon-inline" />,
                onClick: onDuplicateBlock,
                keyboardShortcutLabel: !isInputFocused ? `${modifierKeyLabel} + D` : '',
            },
            {
                label: 'Move Up',
                icon: <ArrowUpIcon className="icon-inline" />,
                onClick: id => onMoveBlock(id, 'up'),
                keyboardShortcutLabel: !isInputFocused ? `${modifierKeyLabel} + ↑` : '',
            },
            {
                label: 'Move Down',
                icon: <ArrowDownIcon className="icon-inline" />,
                onClick: id => onMoveBlock(id, 'down'),
                keyboardShortcutLabel: !isInputFocused ? `${modifierKeyLabel} + ↓` : '',
            },
            {
                label: 'Delete',
                icon: <DeleteIcon className="icon-inline" />,
                onClick: onDeleteBlock,
                keyboardShortcutLabel: !isInputFocused ? `${modifierKeyLabel} + ⌫` : '',
            },
        ]
    }, [isReadOnly, isInputFocused, modifierKeyLabel, onMoveBlock, onDeleteBlock, onDuplicateBlock])
