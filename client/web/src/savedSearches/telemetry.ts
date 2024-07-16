import type { TelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import type { SavedSearchFields } from '../graphql-operations'
import { namespaceTelemetryMetadata } from '../namespaces/telemetry'

export function telemetryRecordSavedSearchViewSearchResults(
    telemetryRecorder: TelemetryRecorder,
    savedSearch: Pick<SavedSearchFields, 'query' | 'viewerCanAdminister'> & {
        owner?: Pick<SavedSearchFields['owner'], '__typename'>
    },
    onPage: 'List' | 'Form' | 'Detail'
): void {
    const metadata: { [key: string]: number } = {
        ...(savedSearch.owner ? namespaceTelemetryMetadata(savedSearch.owner) : undefined),
        viewerCanAdminister: savedSearch.viewerCanAdminister ? 1 : 0,
        queryLength: savedSearch.query.length,
        [`on${onPage}`]: 1,
    }
    telemetryRecorder.recordEvent('savedSearches', 'viewSearchResults', {
        metadata,
    })
}
