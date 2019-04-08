import { isPrivateRepository } from '../util/context'

/**
 * RequestContext tells us if the request is specific to a certain repo and, if so, whether that repo is private.
 */
export interface RequestContext {
    privateRepository: boolean
}

/**
 * getContext takes in some values that you want to be set for a RequestContext
 * and then sets the rest to the default values and returns the rest. You should use
 * this to create a RequestContext if you're setting none or some of the values
 * of the context for a given request.
 */
export function getContext(): RequestContext {
    return { privateRepository: isPrivateRepository() }
}
