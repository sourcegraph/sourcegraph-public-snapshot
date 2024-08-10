// We want to limit the number of imported modules as much as possible
/* eslint-disable no-restricted-imports */

export { parseSearchURL, type ParsedSearchURL } from '@sourcegraph/web/src/search/index'
export { createSuggestionsSource } from '@sourcegraph/web/src/search/input/suggestions'
export { CachedAsyncCompletionSource, type CompletionResult } from '@sourcegraph/web/src/search/autocompletion/source'

export { syntaxHighlight } from '@sourcegraph/web/src/repo/blob/codemirror/highlight'
export { linkify } from '@sourcegraph/web/src/repo/blob/codemirror/links'
export { createCodeIntelExtension } from '@sourcegraph/web/src/repo/blob/codemirror/codeintel/extension'
export type { TooltipViewOptions } from '@sourcegraph/web/src/repo/blob/codemirror/codeintel/api'
export { debugOccurrences } from '@sourcegraph/web/src/repo/blob/codemirror/codeintel/debugOccurrences'
export {
    codeGraphData,
    type CodeGraphData,
    type IndexedCodeGraphData,
} from '@sourcegraph/web/src/repo/blob/codemirror/codeintel/occurrences'

export { positionToOffset, locationToURL } from '@sourcegraph/web/src/repo/blob/codemirror/utils'
export { lockFirstVisibleLine } from '@sourcegraph/web/src/repo/blob/codemirror/lock-line'
export { syncSelection } from '@sourcegraph/web/src/repo/blob/codemirror/codeintel/token-selection'
export {
    showTemporaryTooltip,
    temporaryTooltip,
} from '@sourcegraph/web/src/repo/blob/codemirror/tooltips/TemporaryTooltip'
export type {
    CodeIntelAPIConfig,
    Definition,
    GoToDefinitionOptions,
    DocumentInfo,
} from '@sourcegraph/web/src/repo/blob/codemirror/codeintel/api'
export { type BlameHunk, type BlameHunkData, fetchBlameHunksMemoized } from '@sourcegraph/web/src/repo/blame/shared'
export { blameData, showBlame } from '@sourcegraph/web/src/repo/blob/codemirror/blame-decorations'
export {
    search,
    type SearchPanelView,
    type SearchPanelViewCreationOptions,
    type SearchPanelState,
} from '@sourcegraph/web/src/repo/blob/codemirror/search'
export {
    selectableLineNumbers,
    type SelectedLineRange,
    setSelectedLines,
} from '@sourcegraph/web/src/repo/blob/codemirror/linenumbers'
export { hideEmptyLastLine } from '@sourcegraph/web/src/repo/blob/codemirror/eof'
export { isValidLineRange } from '@sourcegraph/web/src/repo/blob/codemirror/utils'
export { blobPropsFacet } from '@sourcegraph/web/src/repo/blob/codemirror'
export {
    defaultSearchModeFromSettings,
    defaultPatternTypeFromSettings,
    showQueryExamplesForKeywordSearch,
} from '@sourcegraph/web/src/util/settings'

export type { FeatureFlagName } from '@sourcegraph/web/src/featureFlags/featureFlags'

export { parseBrowserRepoURL, getURLToFileCommit } from '@sourcegraph/web/src/util/url'
export type { EditorSettings, EditorReplacements } from '@sourcegraph/web/src/open-in-editor/editor-settings'
export { type Editor, getEditor, supportedEditors } from '@sourcegraph/web/src/open-in-editor/editors'
export {
    buildRepoBaseNameAndPath,
    isProjectPathValid,
    getProjectPath,
    buildEditorUrl,
} from '@sourcegraph/web/src/open-in-editor/build-url'
