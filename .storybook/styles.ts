// Set commonly used CSS variables that many components assume exist. This is better than importing
// all of (e.g.) SourcegraphWebApp.scss, because those styles are only applied for the web app and
// would be misleading to use for browser extension and shared component storybooks.
if (document && document.documentElement) {
    // It's not necessary to define all CSS variables here or to use the precise values from our CSS.
    // These are just used for storybooks.
    const CSS_VARS = {
        '--secondary': '#777777',
        '--text-muted': '#bbbbbb',
        '--primary': '#1c7ed6',
    }
    for (const name of Object.keys(CSS_VARS)) {
        document.documentElement.style.setProperty(name, CSS_VARS[name])
    }
}
