import classNames from 'classnames'
import React from 'react'

import { Menu, MenuButton } from '@sourcegraph/wildcard'

import { SourceLocationSetViewModeActionMenuItems } from '../../../../../repo/actions/source-location-set-view-mode-action/SourceLocationSetViewModeAction'
import { TreeOrComponentViewOptionsProps } from '../../../contributions/tree/TreeOrComponent'

interface Props
    extends Pick<TreeOrComponentViewOptionsProps, 'treeOrComponentViewMode' | 'treeOrComponentViewModeURL'> {
    buttonClassName?: string
}

// TODO(sqs): for clarity, instead make this a dropdown "Show info from <component>"
export const SourceLocationSetSelectMenu: React.FunctionComponent<Props> = ({
    treeOrComponentViewMode,
    treeOrComponentViewModeURL,
    buttonClassName,
}) => (
    <Menu>
        <MenuButton variant="secondary" className={classNames('bg-transparent border-0', buttonClassName)}>
            <span aria-hidden={true}>▾</span>
        </MenuButton>
        <SourceLocationSetViewModeActionMenuItems
            treeOrComponentViewMode={treeOrComponentViewMode}
            treeOrComponentViewModeURL={treeOrComponentViewModeURL}
        />
    </Menu>
)
