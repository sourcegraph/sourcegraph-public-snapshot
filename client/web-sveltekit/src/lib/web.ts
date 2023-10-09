// We want to limit the number of imported modules as much as possible
/* eslint-disable no-restricted-imports */

export { parseSearchURL } from '@sourcegraph/web/src/search/index'
export { createSuggestionsSource } from '@sourcegraph/web/src/search/input/suggestions'
export { replaceRevisionInURL } from '@sourcegraph/web/src/util/url'

export { syntaxHighlight } from '@sourcegraph/web/src/repo/blob/codemirror/highlight'
export { buildLinks } from '@sourcegraph/web/src/repo/blob/codemirror/links'
export { codeIntelAPI } from '@sourcegraph/web/src/repo/blob/codemirror/token-selection/extension'
export {
    selectableLineNumbers,
    type SelectedLineRange,
    setSelectedLines,
} from '@sourcegraph/web/src/repo/blob/codemirror/linenumbers'
export { isValidLineRange } from '@sourcegraph/web/src/repo/blob/codemirror/utils'
export { blobPropsFacet } from '@sourcegraph/web/src/repo/blob/codemirror'
export { defaultSearchModeFromSettings } from '@sourcegraph/web/src/util/settings'

export type { FeatureFlagName } from '@sourcegraph/web/src/featureFlags/featureFlags'
