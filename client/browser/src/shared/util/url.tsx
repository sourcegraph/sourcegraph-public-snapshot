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
    ctx: RepoSpec & RevSpec & FileSpec & Partial<PositionSpec> & Partial<ViewStateSpec>
): string {
    const url = repoUrlCache[ctx.repoName] || sourcegraphUrl
    return `${url.replace(/\/$/, '')}/${toPrettyBlobURL(ctx)}`
}
