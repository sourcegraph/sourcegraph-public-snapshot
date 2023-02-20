import { SourcegraphContext } from '../jscontext'

/**
 * @param isSourcegraphDotCom whether the instance is Sourcegraph.com
 * @param allowSignup whether the instance has built-in signup enabled
 * @param accessRequestsFeatureEnabled whether the access requests experimental feature is explicitly disabled
 * @returns whether the access request feature should be enabled/allowed
 */
export function checkIsRequestAccessAllowed(
    isSourcegraphDotCom: boolean,
    allowSignup: boolean,
    accessRequestsFeatureEnabled: SourcegraphContext['experimentalFeatures']['accessRequests.enabled']
): boolean {
    return !isSourcegraphDotCom && !allowSignup && accessRequestsFeatureEnabled !== false
}
