import { Alert, Text } from '@sourcegraph/wildcard'

import type { EnterprisePortalEnvironment } from './enterpriseportal'

export interface EnterprisePortalEnvWarningProps {
    env: EnterprisePortalEnvironment
    /**
     * For example, 'creating a subscription' - this will be inserted into the
     * warning text.
     */
    actionText: string

    className?: string
}

/**
 * Displays a warning about the user's action if the selected env is not 'prod'.
 */
export const EnterprisePortalEnvWarning: React.FunctionComponent<EnterprisePortalEnvWarningProps> = ({
    env,
    actionText,
    className,
}) => {
    if (env === 'prod') {
        return null
    }
    return (
        <Alert variant="danger" className={className}>
            <Text>
                You are {actionText} for the <strong>{env}</strong> Enterprise Portal deployment. Everything you do here
                will only be visible to non-production environments that connect to this specific Enterprise Portal
                deployment.
            </Text>
            <Text className="mb-0">
                If you are {actionText} for a customer, select <strong>Production</strong> from the "Enterprise Portal"
                dropdown in the top right
            </Text>
        </Alert>
    )
}
