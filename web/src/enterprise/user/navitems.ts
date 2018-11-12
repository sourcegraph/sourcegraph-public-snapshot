import SubscriptionsIcon from 'mdi-react/SubscriptionsIcon'
import { userAreaHeaderNavItems } from '../../user/area/navitems'
import { UserAreaHeaderNavItem } from '../../user/area/UserAreaHeader'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'

export const enterpriseUserAreaHeaderNavItems: ReadonlyArray<UserAreaHeaderNavItem> = [
    ...userAreaHeaderNavItems,
    {
        label: 'Subscriptions',
        icon: SubscriptionsIcon,
        to: '/subscriptions',
        condition: ({ user }) => SHOW_BUSINESS_FEATURES && user.viewerCanAdminister,
    },
]
