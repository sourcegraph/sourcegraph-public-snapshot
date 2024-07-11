// These are currently used by the web app and because the web-sveltekit app uses
// code from the web app these need to be defined here as well.
interface PageError {
    statusCode: number
    statusText: string
    error: string
    errorID: string
}

interface Window {
    pageError?: PageError
    context: import('@sourcegraph/web/src/jscontext').SourcegraphContext
}

// Importing highlight.js/lib/core or a language (highlight.js/lib/languages/*) results in
// a compiler error about not being able to find the types. Adding this declaration fixes it.
declare module 'highlight.js/lib/core' {
    export * from 'highlight.js'
}
