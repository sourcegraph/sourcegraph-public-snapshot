import { AuthenticatedUser } from '../../../auth'

interface IsSentinelEnabledProps {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser | null
}

export const isSentinelEnabled = (props: IsSentinelEnabledProps): boolean => {
    const { isSourcegraphDotCom, authenticatedUser } = props

    if (authenticatedUser) {
        return isSourcegraphDotCom && authenticatedUser.siteAdmin
    }

    return false
}
