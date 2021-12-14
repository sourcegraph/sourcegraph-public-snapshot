import { Menu as ReachMenu, MenuProps as ReachMenuProps } from '@reach/menu-button'
import React from 'react'

export type MenuProps = ReachMenuProps

export const Menu: React.FunctionComponent<ReachMenuProps> = props => <ReachMenu {...props} />
