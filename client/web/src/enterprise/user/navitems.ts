import { userAreaHeaderNavItems } from '../../user/area/navitems'
import { UserAreaTabsNavItem } from '../../user/area/UserAreaTabs'
import { enterpriseNamespaceAreaHeaderNavItems } from '../namespaces/navitems'

export const enterpriseUserAreaTabsNavItems: readonly UserAreaTabsNavItem[] = [
    ...userAreaHeaderNavItems,
    ...enterpriseNamespaceAreaHeaderNavItems,
]
