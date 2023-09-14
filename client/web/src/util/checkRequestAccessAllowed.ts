import type { SourcegraphContext } from '../jscontext'

/**
 * @returns whether the access request feature is allowed or not
 */
export function checkRequestAccessAllowed({
    sourcegraphDotComMode,
    allowSignup,
    authAccessRequest,
}: Pick<SourcegraphContext, 'sourcegraphDotComMode' | 'allowSignup' | 'authAccessRequest'>): boolean {
    return !sourcegraphDotComMode && !allowSignup && authAccessRequest?.enabled !== false
}
