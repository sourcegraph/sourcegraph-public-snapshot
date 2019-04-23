import {
    FileSpec,
    PositionSpec,
    RepoSpec,
    RevSpec,
    toPrettyBlobURL,
    ViewStateSpec,
} from '../../../../../shared/src/util/url'
import { sourcegraphUrl } from './context'

/**
 * Returns an absolute URL to the blob (file) on the Sourcegraph instance.
 */
export function toAbsoluteBlobURL(
    ctx: RepoSpec & RevSpec & FileSpec & Partial<PositionSpec> & Partial<ViewStateSpec>
): string {
    // toPrettyBlobURL() always returns an URL starting with a forward slash,
    // no need to add one here
    return `${sourcegraphUrl.replace(/\/$/, '')}${toPrettyBlobURL(ctx)}`
}
