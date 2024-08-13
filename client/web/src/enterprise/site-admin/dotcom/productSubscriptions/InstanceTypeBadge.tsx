import { Badge, type BadgeVariantType } from '@sourcegraph/wildcard'

import { EnterpriseSubscriptionInstanceType } from './enterpriseportalgen/subscriptions_pb'

export interface InstanceTypeBadgeProps {
    instanceType: EnterpriseSubscriptionInstanceType
    className?: string
}

/**
 * Displays instance type in a cute badge with relevant tooltips.
 */
export const InstanceTypeBadge: React.FunctionComponent<InstanceTypeBadgeProps> = ({ instanceType, className }) => {
    let variant: BadgeVariantType = 'outlineSecondary'
    let tooltip = ''
    switch (instanceType) {
        case EnterpriseSubscriptionInstanceType.INTERNAL: {
            tooltip = 'This subscription is for Sourcegraph-internal instances.'
            break
        }
        case EnterpriseSubscriptionInstanceType.PRIMARY: {
            variant = 'primary'
            tooltip = "This subscription is for a customer's primary, production Sourcegraph instance."
            break
        }
        case EnterpriseSubscriptionInstanceType.SECONDARY: {
            variant = 'info'
            tooltip =
                "This subscription is for a customer's secondary Sourcegraph instance, such as one for staging new releases."
            break
        }
    }
    return (
        <Badge variant={variant} small={true} tooltip={tooltip} className={className}>
            {EnterpriseSubscriptionInstanceType[instanceType]}
        </Badge>
    )
}
