import type { FC } from 'react'

import { Menu, MenuButton, MenuList } from '@sourcegraph/wildcard'

import type { RepoHeaderContribution } from './RepoHeader'

interface RepoHeaderContextMenuProps {
    actions: (Pick<RepoHeaderContribution, 'id' | 'priority'> & { element: JSX.Element | null })[]
}

export const RepoHeaderContextMenu: FC<RepoHeaderContextMenuProps> = ({ actions }) => (
    <Menu>
        <MenuButton variant="secondary" outline={true} className="pt-0 pb-0">
            ...
        </MenuButton>

        <MenuList>{actions.map(action => action.element)}</MenuList>
    </Menu>
)
