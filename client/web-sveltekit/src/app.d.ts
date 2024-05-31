// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
/// <reference types="@sveltejs/kit" />

import 'unplugin-icons/types/svelte'

declare global {
    namespace App {
        interface PageData {
            // Used by the repository pages to control the history panel
            enableInlineDiff?: boolean
            enableViewAtCommit?: boolean
        }
    }
}
