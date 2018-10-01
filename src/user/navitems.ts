import { userAreaHeaderNavItems } from '@sourcegraph/webapp/dist/user/area/navitems'
import { UserAreaHeaderNavItem } from '@sourcegraph/webapp/dist/user/area/UserAreaHeader'
import SubscriptionsIcon from 'mdi-react/SubscriptionsIcon'
import { USE_DOTCOM_BUSINESS } from '../dotcom/productSubscriptions/features'

export const enterpriseUserAreaHeaderNavItems: ReadonlyArray<UserAreaHeaderNavItem> = [
    ...userAreaHeaderNavItems,
    {
        label: 'Subscriptions',
        icon: SubscriptionsIcon,
        to: '/subscriptions',
        condition: () => USE_DOTCOM_BUSINESS,
    },
]
