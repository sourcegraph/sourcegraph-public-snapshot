import classNames from 'classnames'
import React, { useRef } from 'react'

import { Menu, MenuButton, MenuPopover } from '@sourcegraph/wildcard'

import { SourceLocationSetViewModeActionMenuItems } from '../../../../../repo/actions/source-location-set-view-mode-action/SourceLocationSetViewModeAction'
import { positionBottomRight } from '../../../../insights/components/context-menu/utils'
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
}) => {
    const targetButtonReference = useRef<HTMLButtonElement>(null)

    return (
        <Menu>
            <MenuButton
                variant="secondary"
                className={classNames('bg-transparent border-0', buttonClassName)}
                ref={targetButtonReference}
            >
                <span aria-hidden={true}>▾</span>
            </MenuButton>
            <MenuPopover position={positionBottomRight}>
                <SourceLocationSetViewModeActionMenuItems
                    treeOrComponentViewMode={treeOrComponentViewMode}
                    treeOrComponentViewModeURL={treeOrComponentViewModeURL}
                />
            </MenuPopover>
        </Menu>
    )
}
