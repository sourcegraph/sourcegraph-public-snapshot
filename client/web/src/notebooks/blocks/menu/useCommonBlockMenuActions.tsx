import { useMemo } from 'react'

import { mdiContentDuplicate, mdiArrowUp, mdiArrowDown, mdiDelete, mdiPlusCircleOutline } from '@mdi/js'

import { isMacPlatform as isMacPlatformFunc } from '@sourcegraph/common'
import { Icon } from '@sourcegraph/wildcard'

import type { BlockProps } from '../..'
import { useIsBlockInputFocused } from '../useIsBlockInputFocused'
import { useModifierKeyLabel } from '../useModifierKeyLabel'

import type { BlockMenuAction } from './NotebookBlockMenu'

export const useCommonBlockMenuActions = ({
    id,
    isReadOnly,
    onMoveBlock,
    onDeleteBlock,
    onNewBlock,
    onDuplicateBlock,
}: Pick<
    BlockProps,
    'id' | 'isReadOnly' | 'onDeleteBlock' | 'onDuplicateBlock' | 'onMoveBlock' | 'onNewBlock'
>): BlockMenuAction[] => {
    const isMacPlatform = useMemo(() => isMacPlatformFunc(), [])
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
                icon: <Icon aria-hidden={true} svgPath={mdiContentDuplicate} />,
                onClick: onDuplicateBlock,
                keyboardShortcutLabel: !isInputFocused ? `${modifierKeyLabel} + D` : '',
            },
            {
                type: 'button',
                label: 'Move Up',
                icon: <Icon aria-hidden={true} svgPath={mdiArrowUp} />,
                onClick: id => onMoveBlock(id, 'up'),
                keyboardShortcutLabel: !isInputFocused ? `${modifierKeyLabel} + ↑` : '',
            },
            {
                type: 'button',
                label: 'Move Down',
                icon: <Icon aria-hidden={true} svgPath={mdiArrowDown} />,
                onClick: id => onMoveBlock(id, 'down'),
                keyboardShortcutLabel: !isInputFocused ? `${modifierKeyLabel} + ↓` : '',
            },
            {
                type: 'button',
                label: 'Delete',
                icon: <Icon aria-hidden={true} svgPath={mdiDelete} />,
                onClick: onDeleteBlock,
                keyboardShortcutLabel: !isInputFocused ? (isMacPlatform ? '⌘ + ⌫' : 'Del') : '',
            },
            {
                type: 'button',
                label: 'Add Block',
                icon: <Icon aria-hidden={true} svgPath={mdiPlusCircleOutline} />,
                onClick: onNewBlock,
                keyboardShortcutLabel: `${modifierKeyLabel} + ⇧ + ↵`,
            },
        ]
    }, [
        isReadOnly,
        isMacPlatform,
        isInputFocused,
        modifierKeyLabel,
        onMoveBlock,
        onDeleteBlock,
        onDuplicateBlock,
        onNewBlock,
    ])
}
