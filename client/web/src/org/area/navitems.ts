import AccountMultipleIcon from 'mdi-react/AccountMultipleIcon'
import CogOutlineIcon from 'mdi-react/CogOutlineIcon'
import FeatureSearchOutlineIcon from 'mdi-react/FeatureSearchOutlineIcon'

import { namespaceAreaHeaderNavItems } from '../../namespaces/navitems'

import { OrgAreaHeaderNavItem } from './OrgHeader'

export const orgAreaHeaderNavItems: readonly OrgAreaHeaderNavItem[] = [
    {
        to: '/settings/members',
        label: 'Members',
        icon: AccountMultipleIcon,
        isActive: (_match, location) => location.pathname.includes('members'),
        condition: ({ org: { viewerCanAdminister }, newMembersInviteEnabled }) =>
            viewerCanAdminister && newMembersInviteEnabled,
    },
    {
        to: '/settings',
        label: 'Settings',
        icon: CogOutlineIcon,
        isActive: (_match, location, context) =>
            context.newMembersInviteEnabled
                ? location.pathname.includes('settings') && !location.pathname.includes('members')
                : location.pathname.includes('settings'),
        condition: ({ org: { viewerCanAdminister } }) => viewerCanAdminister,
    },
    {
        to: '/searches',
        label: 'Saved searches',
        icon: FeatureSearchOutlineIcon,
        condition: ({ org: { viewerCanAdminister } }) => viewerCanAdminister,
    },
    ...namespaceAreaHeaderNavItems,
]
