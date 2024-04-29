import { EventLogger } from '@sourcegraph/shared/src/telemetry/web/eventLogger'

class SvelteTelemetry extends EventLogger {
    public override logViewEvent(event: string): void {
        super.logViewEvent(event, { isSveltePrototype: true })
    }

    public override log(eventName: string, eventProperties?: any, publicArgument?: any): void {
        super.log(
            eventName,
            { ...eventProperties, isSveltePrototype: true },
            { ...publicArgument, isSveltePrototype: true }
        )
    }
}

export const SVELTE_LOGGER = new SvelteTelemetry()

// These events are minimal set of telemetry events that we
// use in react version, note that names should be identical
// with event names that we use in react
export enum SVELTE_TELEMETRY_EVENTS {
    // Note that prefix 'View' will be added by EventLogger
    ViewHomePage = 'Home',
    ViewSearchResultsPage = 'SearchResults',
    ViewRepositoryPage = 'Repository',
    ViewBlobPage = 'Blob',

    SearchSubmit = 'SearchSubmitted',
    SearchResultClick = 'SearchResultClicked',
    ShowHistoryPanel = 'ShowHistoryPanel',
    HideHistoryPanel = 'HideHistoryPanel',
    CodeCopied = 'CodeCopied',
    GoToCodeHost = 'GoToCodeHostClicked',
    GitBlameEnabled = 'GitBlameEnabled',
    SelectSearchFilter = 'SearchFiltersSelectFilter',
}

export const codeCopiedEvent = (page: string): [string, { page: string }, { page: string }] => [
    SVELTE_TELEMETRY_EVENTS.CodeCopied,
    { page },
    { page },
]
