import SubscriptionsIcon from 'mdi-react/SubscriptionsIcon'
import { userAreaHeaderNavItems } from '../../../packages/webapp/src/user/area/navitems'
import { UserAreaHeaderNavItem } from '../../../packages/webapp/src/user/area/UserAreaHeader'
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
