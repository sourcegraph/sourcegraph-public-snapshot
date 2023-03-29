import { SourcegraphContext } from '../jscontext'

/**
 * @returns whether the access request feature is allowed or not
 */
export function checkRequestAccessAllowed(
    context: Pick<SourcegraphContext, 'sourcegraphDotComMode' | 'allowSignup' | 'authAccessRequest'>
): boolean {
    return !context.sourcegraphDotComMode && !context.allowSignup && !context.authAccessRequest?.disabled
}
