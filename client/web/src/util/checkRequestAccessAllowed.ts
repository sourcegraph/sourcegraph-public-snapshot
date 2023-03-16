import { SourcegraphContext } from '../jscontext'

/**
 * @param isSourcegraphDotCom whether the instance is Sourcegraph.com
 * @param allowSignup whether the instance has built-in signup enabled
 * @param accessRequestsFeatureEnabled whether the access requests experimental feature is explicitly disabled
 * @returns whether the access request feature is allowed or not
 */
export function checkRequestAccessAllowed(
    isSourcegraphDotCom: boolean,
    allowSignup: boolean,
    accessRequestsFeatures: Pick<SourcegraphContext['experimentalFeatures'], 'accessRequest.enabled'> | undefined
): boolean {
    return !isSourcegraphDotCom && !allowSignup && accessRequestsFeatures?.['accessRequest.enabled'] !== false
}
