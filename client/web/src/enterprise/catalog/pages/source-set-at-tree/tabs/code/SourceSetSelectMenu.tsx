import classNames from 'classnames'
import React from 'react'

import { Menu, MenuButton } from '@sourcegraph/wildcard'

import { SourceSetViewModeActionMenuItems } from '../../../../../../repo/actions/source-set-view-mode-action/SourceSetViewModeAction'
import { SourceSetAtTreeViewOptionsProps } from '../../../contributions/tree/useSourceSetAtTreeViewOptions'

interface Props
    extends Pick<SourceSetAtTreeViewOptionsProps, 'sourceSetAtTreeViewMode' | 'sourceSetAtTreeViewModeURL'> {
    buttonClassName?: string
}

// TODO(sqs): for clarity, instead make this a dropdown "Show info from <component>"
export const SourceSetSelectMenu: React.FunctionComponent<Props> = ({
    sourceSetAtTreeViewMode,
    sourceSetAtTreeViewModeURL,
    buttonClassName,
}) => (
    <Menu>
        <MenuButton variant="secondary" className={classNames('bg-transparent border-0', buttonClassName)}>
            <span aria-hidden={true}>▾</span>
        </MenuButton>
        <SourceSetViewModeActionMenuItems
            sourceSetAtTreeViewMode={sourceSetAtTreeViewMode}
            sourceSetAtTreeViewModeURL={sourceSetAtTreeViewModeURL}
        />
    </Menu>
)
