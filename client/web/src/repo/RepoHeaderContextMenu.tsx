import type { FC } from 'react'

import { Menu, MenuButton, MenuList, Position } from '@sourcegraph/wildcard'

import type { RepoHeaderContribution } from './RepoHeader'

interface RepoHeaderContextMenuProps {
    actions: (Pick<RepoHeaderContribution, 'id' | 'priority'> & { element: JSX.Element | null })[]
}

export const RepoHeaderContextMenu: FC<RepoHeaderContextMenuProps> = ({ actions }) => (
    <Menu>
        <MenuButton variant="secondary" outline={true} className="py-0">
            &hellip;
        </MenuButton>

        <MenuList position={Position.bottom}>{actions.map(action => action.element)}</MenuList>
    </Menu>
)
