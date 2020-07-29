/**
 * Whether the experimental code modification feature is enabled.
 *
 * To enable this, run `localStorage.codemodExp=true;location.reload()` in your browser's JavaScript
 * console.
 */
export const USE_CODEMOD = localStorage.getItem('codemodExp') !== null
