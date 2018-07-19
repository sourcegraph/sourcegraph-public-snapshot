import * as GQL from '../backend/graphqlschema'
import { USE_PLATFORM } from '../cxp/CXPEnvironment'

const PLATFORM_TAG = 'platform'

/**
 * Whether the platform (extensions, the extension registry, CXP, and WebSockets) should be enabled for the user
 * (who is assumed to be the current viewer).
 *
 * On local dev instances, run `localStorage.platform=true;location.reload()` to enable this.
 *
 * The server (site config experimentalFeatures.platform value), user (DB users.tags array containing "platform"),
 * and browser (localStorage.platform) feature flags must be enabled for this to be true.
 *
 * On non-Sourcegraph.com instances with experimentalFeatures.platform true in site config, the "platform" tag is
 * automatically added to users.
 *
 */
export function platformEnabled(user: Pick<GQL.IUser, 'tags'>): boolean {
    return !!window.context.platformEnabled && hasTag(user, PLATFORM_TAG) && USE_PLATFORM
}

function hasTag(user: Pick<GQL.IUser, 'tags'>, tag: string): boolean {
    return user.tags.includes(tag)
}
