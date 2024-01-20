import CogOutlineIcon from 'mdi-react/CogOutlineIcon'
import FeatureSearchOutlineIcon from 'mdi-react/FeatureSearchOutlineIcon'

import { namespaceAreaHeaderNavItems } from '../../namespaces/navitems'
import { isCodyOnlyLicense } from '../../util/license'

import type { OrgAreaHeaderNavItem } from './OrgHeader'

const disableCodeSearchFeatures = isCodyOnlyLicense()

export const orgAreaHeaderNavItems: readonly OrgAreaHeaderNavItem[] = [
    {
        to: '/settings',
        label: 'Settings',
        icon: CogOutlineIcon,
        condition: ({ org: { viewerCanAdminister } }) => viewerCanAdminister,
    },
    {
        to: '/searches',
        label: 'Saved searches',
        icon: FeatureSearchOutlineIcon,
        condition: ({ org: { viewerCanAdminister } }) => !disableCodeSearchFeatures && viewerCanAdminister,
    },
    ...namespaceAreaHeaderNavItems,
]
