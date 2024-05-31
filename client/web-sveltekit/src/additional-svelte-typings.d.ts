// todo(@fkling): Consider to extend HTMLInputAttributes directly after upgrading to svelte@4.2, svelte-check@3.5
// See: https://svelte.dev/docs/typescript#enhancing-built-in-dom-types
declare namespace svelteHTML {
    interface HTMLAttributes<T> {
        /**
         * This is used by CodeMirror to identify the input field in the search panel.
         */
        'main-field'?: boolean
    }
}
