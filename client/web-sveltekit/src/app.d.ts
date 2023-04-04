// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
/// <reference types="@sveltejs/kit" />

import type { ErrorLike } from '$lib/common'
import type { ResolvedRevision, Repo } from '$lib/web'

// and what to do when importing types
declare namespace App {
    interface PageData {
        resolvedRevision?: (ResolvedRevision & Repo) | ErrorLike
    }
}
