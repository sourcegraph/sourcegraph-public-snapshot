import classNames from 'classnames'
import React from 'react'

import { Menu, MenuButton } from '@sourcegraph/wildcard'

import { SourceSetViewModeActionMenuItems } from '../../../../../repo/actions/source-set-view-mode-action/SourceSetViewModeAction'
import { TreeOrComponentViewOptionsProps } from '../../../contributions/tree/TreeOrComponent'

interface Props
    extends Pick<TreeOrComponentViewOptionsProps, 'treeOrComponentViewMode' | 'treeOrComponentViewModeURL'> {
    buttonClassName?: string
}

// TODO(sqs): for clarity, instead make this a dropdown "Show info from <component>"
export const SourceSetSelectMenu: React.FunctionComponent<Props> = ({
    treeOrComponentViewMode,
    treeOrComponentViewModeURL,
    buttonClassName,
}) => (
    <Menu>
        <MenuButton variant="secondary" className={classNames('bg-transparent border-0', buttonClassName)}>
            <span aria-hidden={true}>▾</span>
        </MenuButton>
        <SourceSetViewModeActionMenuItems
            treeOrComponentViewMode={treeOrComponentViewMode}
            treeOrComponentViewModeURL={treeOrComponentViewModeURL}
        />
    </Menu>
)
