type Entry = 'src/enterprise/main' | 'src/enterprise/embed/embedMain' | 'src/enterprise/app/main'

export interface WebBuildManifest {
    /** Base URL for asset paths. */
    url?: string

    /**
     * A map of entrypoint (such as "src/enterprise/main" with no extension) to its JavaScript and
     * CSS assets.
     */
    assets: Partial<Record<Entry, { js: string; css?: string }>>

    /** Additional HTML <script> tags to inject in dev mode. */
    devInjectHTML?: string
}
