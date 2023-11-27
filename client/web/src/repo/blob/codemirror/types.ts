/**
 * Blob UI search panels mode variations. The code mirror blob UI has
 * inline search UI (you can open it with cmd + F), with which you can
 * found any arbitrary matches based on this search value.
 */
export enum SearchPanelViewMode {
    /**
     * Plain inline search mode,
     * renders search input, matches count and its controls,
     * command F controls and info tooltip.
     */
    FullSearch = 'full-search',

    /**
     * This mode is experimental and was added only for blob UI
     * file preview case. In file preview we shouldn't have UI like
     * command F controls and info tooltip. In matches-only mode we
     * render only search input and matches UI.
     */
    MatchesOnly = 'matches-only',
}
