// Importing highlight.js/lib/core or a language (highlight.js/lib/languages/*) results in
// a compiler error about not being able to find the types. Adding this declaration fixes it.
declare module 'highlight.js/lib/core' {
    export * from 'highlight.js'
}
