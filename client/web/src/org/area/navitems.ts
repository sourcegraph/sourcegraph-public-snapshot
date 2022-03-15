import AccountMultipleIcon from 'mdi-react/AccountMultipleIcon'
import CogOutlineIcon from 'mdi-react/CogOutlineIcon'
import FeatureSearchOutlineIcon from 'mdi-react/FeatureSearchOutlineIcon'
import PlayCircleOutlineIcon from 'mdi-react/PlayCircleOutlineIcon'

import { namespaceAreaHeaderNavItems } from '../../namespaces/navitems'
import { showGetStartPage } from '../openBeta/GettingStarted'

import { OrgGetStartedInfo } from './OrgArea'
import { OrgAreaHeaderNavItem } from './OrgHeader'

const calculateLeftGetStartedSteps = (info: OrgGetStartedInfo): number => {
    let leftSteps = 1
    if (info.invitesCount === 0 && info.membersCount < 2) {
        leftSteps += 1
    }
    if (info.reposCount === 0) {
        leftSteps += 1
    }
    if (info.servicesCount === 0) {
        leftSteps += 1
    }

    return leftSteps
}

export const orgAreaHeaderNavItems: readonly OrgAreaHeaderNavItem[] = [
    {
        to: '/getstarted',
        label: 'Get started',
        dynamicLabel: ({ getStartedInfo: info }) => `Get started ${calculateLeftGetStartedSteps(info)}`,
        icon: PlayCircleOutlineIcon,
        isActive: (_match, location) => location.pathname.includes('getstarted'),
        condition: showGetStartPage,
    },
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
