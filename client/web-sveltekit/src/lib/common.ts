// We want to limit the number of imported modules as much as possible
/* eslint-disable no-restricted-imports */
export type { ErrorLike } from '@sourcegraph/common/src/errors/types'
export { isErrorLike } from '@sourcegraph/common/src/errors/utils'
export { createAggregateError, asError } from '@sourcegraph/common/src/errors/errors'
export { memoizeObservable, resetAllMemoizationCaches } from '@sourcegraph/common/src/util/rxjs/memoizeObservable'
export {
    encodeURIPathComponent,
    toPositionOrRangeQueryParameter,
    addLineRangeQueryParameter,
    formatSearchParameters,
} from '@sourcegraph/common/src/util/url'
export { pluralize, numberWithCommas } from '@sourcegraph/common/src/util/strings'
export { renderMarkdown } from '@sourcegraph/common/src/util/markdown/markdown'
export { highlightNodeMultiline, highlightNode } from '@sourcegraph/common/src/util/highlightNode'
export { logger } from '@sourcegraph/common/src/util/logger'
