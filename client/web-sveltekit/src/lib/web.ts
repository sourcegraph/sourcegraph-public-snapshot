// We want to limit the number of imported modules as much as possible
/* eslint-disable no-restricted-imports */

export { parseSearchURL } from '@sourcegraph/web/src/search/index'
export { createSuggestionsSource } from '@sourcegraph/web/src/search/input/suggestions'
export { replaceRevisionInURL } from '@sourcegraph/web/src/util/url'

export { syntaxHighlight } from '@sourcegraph/web/src/repo/blob/codemirror/highlight'
export { linkify } from '@sourcegraph/web/src/repo/blob/codemirror/links'
export { createCodeIntelExtension } from '@sourcegraph/web/src/repo/blob/codemirror/codeintel/extension'
export type { TooltipViewOptions } from '@sourcegraph/web/src/repo/blob/codemirror/codeintel/api'
export { positionToOffset, locationToURL } from '@sourcegraph/web/src/repo/blob/codemirror/utils'
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
export {
    selectableLineNumbers,
    type SelectedLineRange,
    setSelectedLines,
} from '@sourcegraph/web/src/repo/blob/codemirror/linenumbers'
export { isValidLineRange } from '@sourcegraph/web/src/repo/blob/codemirror/utils'
export { blobPropsFacet } from '@sourcegraph/web/src/repo/blob/codemirror'
export { defaultSearchModeFromSettings } from '@sourcegraph/web/src/util/settings'
export { containsTest } from '@sourcegraph/web/src/repo/utils'

export type { FeatureFlagName } from '@sourcegraph/web/src/featureFlags/featureFlags'
