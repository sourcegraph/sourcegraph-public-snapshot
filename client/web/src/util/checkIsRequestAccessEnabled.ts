import { SourcegraphContext } from '../jscontext'

export function checkIsRequestAccessEnabled(
    isSourcegraphDotCom: boolean,
    allowSignup: boolean,
    requestAccessExperimentalFeature: SourcegraphContext['experimentalFeatures']['requestAccess']
): boolean {
    // TODO: check case when allowSignup=true and no seats
    return !isSourcegraphDotCom && !allowSignup && requestAccessExperimentalFeature?.enabled !== false
}
