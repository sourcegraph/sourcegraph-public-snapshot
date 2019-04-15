import {
    FileSpec,
    PositionSpec,
    RepoSpec,
    RevSpec,
    toPrettyBlobURL,
    ViewStateSpec,
} from '../../../../../shared/src/util/url'
import { repoUrlCache, sourcegraphUrl } from './context'

/**
 * Returns an absolute URL to the blob (file) on the Sourcegraph instance.
 */
export function toAbsoluteBlobURL(
    ctx: RepoSpec & RevSpec & FileSpec & Partial<PositionSpec> & Partial<ViewStateSpec>,
    cache = repoUrlCache
): string {
    const url = cache[ctx.repoName] || sourcegraphUrl
    // toPrettyBlobURL() always returns an URL starting with a forward slash,
    // no need to add one here
    return `${url.replace(/\/$/, '')}${toPrettyBlobURL(ctx)}`
}
