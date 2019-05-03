import { isInPage, isPhabricator } from '../../context'
import { parseURL } from '../../libs/github/util'
import { isPrivateRepository } from '../util/context'

// TODO: Make a code host agnostic parseURL so we don't reach into an application directory from the shared directory.
// Let's do this when we make code hosts an interface. Require a `parseURL: () => ParsedRepoURI`
/**
 * RequestContext tells us what repo the request is for and if the request is specific
 * to a certain repo. This information helps the `repoCache` know whether it should
 * cache a url for a repo or if the repoCache should return a cached url or just the
 * current primary url.
 */
export interface RequestContext {
    repoKey: string
    isRepoSpecific: boolean
    privateRepository?: boolean
    /**
     * Requests can provide a blacklist of urls that we don't ever want to try for that given request.
     */
    blacklist?: string[]
}

const defaultContext: RequestContext = {
    repoKey: '',
    isRepoSpecific: true,
}

/**
 * getContext takes in some values that you want to be set for a RequestContext
 * and then sets the rest to the default values and returns the rest. You should use
 * this to create a RequestContext if you're setting none or some of the values
 * of the context for a given request.
 */
export function getContext(ctx: Partial<RequestContext> = {}): RequestContext {
    if (isInPage || isPhabricator) {
        // Phabricator repositories are always considered private
        return { ...defaultContext, ...ctx, privateRepository: true }
    }

    const repoKey = ctx.repoKey

    const out = { ...defaultContext, ...ctx }

    if (!repoKey && out.isRepoSpecific) {
        const { repoName } = parseURL()
        out.repoKey = repoName
    }

    if (out.privateRepository === undefined) {
        out.privateRepository = isPrivateRepository()
    }

    return out
}
