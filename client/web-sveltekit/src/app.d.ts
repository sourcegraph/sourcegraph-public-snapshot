// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
/// <reference types="@sveltejs/kit" />

declare global {
    namespace App {
        interface PageData {}
    }
}

// Importing highlight.js/lib/core or a language (highlight.js/lib/languages/*) results in
// a compiler error about not being able to find the types. Adding this declaration fixes it.
declare module 'highlight.js/lib/core' {
    export * from 'highlight.js'
}
