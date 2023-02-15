import { SourcegraphContext } from '../jscontext'

export function checkIsRequestAccessAllowed(
    isSourcegraphDotCom: boolean,
    allowSignup: boolean,
    requestAccessExperimentalFeature: SourcegraphContext['experimentalFeatures']['requestAccess']
): boolean {
    // TODO: check case when allowSignup=true and no seats
    return !isSourcegraphDotCom && !allowSignup && requestAccessExperimentalFeature?.enabled !== false
}
