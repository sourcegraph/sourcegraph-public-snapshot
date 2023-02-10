import { SourcegraphContext } from '../jscontext'

export function useShouldShowRequestAccess(
    isSourcegraphDotCom: boolean,
    allowSignup: boolean,
    experimentalFeatures: SourcegraphContext['experimentalFeatures']
): boolean {
    return true
    // TODO: check case when allowSignup=true and no seats
    // return !isSourcegraphDotCom && !allowSignup && experimentalFeatures?.requestAccess?.enabled !== false
}
