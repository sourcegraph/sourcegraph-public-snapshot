import type { AuthenticatedUser } from '../../../auth'

interface IsSentinelEnabledProps {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser | null
}

export const isSentinelEnabled = (props: IsSentinelEnabledProps): boolean => {
    const { authenticatedUser } = props

    if (authenticatedUser) {
        return authenticatedUser.siteAdmin
    }

    return false
}
