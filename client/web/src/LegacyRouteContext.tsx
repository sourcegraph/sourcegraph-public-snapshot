import { createContext, useContext, useCallback } from 'react'

import { Observable } from 'rxjs'

import { isMacPlatform } from '@sourcegraph/common'
import { FetchFileParameters, fetchHighlightedFileLineRanges } from '@sourcegraph/shared/src/backend/file'
import {
    fetchSearchContextBySpec,
    fetchSearchContexts,
    fetchSearchContext,
    getUserSearchContextNamespaces,
    createSearchContext,
    updateSearchContext,
    deleteSearchContext,
    isSearchContextSpecAvailable,
} from '@sourcegraph/shared/src/search'
import { aggregateStreamingSearch } from '@sourcegraph/shared/src/search/stream'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { globbingEnabledFromSettings } from '@sourcegraph/shared/src/util/globbing'

import { BatchChangesProps, isBatchChangesExecutionEnabled } from './batches'
import { CodeIntelligenceProps } from './codeintel'
import { BreadcrumbSetters, BreadcrumbsProps } from './components/Breadcrumbs'
import type { LegacyLayoutProps } from './LegacyLayout'
import { ThemePreferenceProps } from './theme'

export interface LegacyLayoutRouteComponentProps
    extends Omit<LegacyLayoutProps, 'match'>,
        ThemeProps,
        ThemePreferenceProps,
        BreadcrumbsProps,
        BreadcrumbSetters,
        CodeIntelligenceProps,
        BatchChangesProps {
    isSourcegraphDotCom: boolean
    isMacPlatform: boolean
}

interface Props {
    render: (props: LegacyLayoutRouteComponentProps) => JSX.Element
    condition?: (props: LegacyLayoutRouteComponentProps) => boolean
}

/**
 * A wrapper component for React router route entrypoints that still need access to the legacy
 * route context and prop drilling.
 */
export const LegacyRoute = ({ render, condition }: Props): JSX.Element | null => {
    const context = useContext(LegacyRouteContext)
    if (!context) {
        throw new Error('LegacyRoute must be used inside a LegacyRouteContext.Provider')
    }

    if (condition && !condition(context)) {
        return null
    }

    return render(context)
}

export interface LegacyRouteContextProviderProps {
    context: Omit<
        LegacyLayoutRouteComponentProps,
        | 'isMacPlatform'
        | 'isSourcegraphDotCom'
        | 'getUserSearchContextNamespaces'
        | 'fetchSearchContexts'
        | 'fetchSearchContextBySpec'
        | 'fetchSearchContext'
        | 'createSearchContext'
        | 'updateSearchContext'
        | 'deleteSearchContext'
        | 'isSearchContextSpecAvailable'
        | 'globbing'
        | 'streamSearch'
        | 'batchChangesExecutionEnabled'
        | 'fetchHighlightedFileLineRanges'
        | 'batchChangesWebhookLogsEnabled'
    >
}
export const LegacyRouteContextProvider: React.FC<React.PropsWithChildren<LegacyRouteContextProviderProps>> = ({
    children,
    context,
}) => {
    const { settingsCascade, platformContext } = context

    const _fetchHighlightedFileLineRanges = useCallback(
        (parameters: FetchFileParameters, force?: boolean | undefined): Observable<string[][]> =>
            fetchHighlightedFileLineRanges({ ...parameters, platformContext }, force),
        [platformContext]
    )

    const legacyContext = {
        ...context,
        isMacPlatform: isMacPlatform(),
        isSourcegraphDotCom: window.context.sourcegraphDotComMode,
        getUserSearchContextNamespaces,
        fetchSearchContexts,
        fetchSearchContextBySpec,
        fetchSearchContext,
        createSearchContext,
        updateSearchContext,
        deleteSearchContext,
        isSearchContextSpecAvailable,
        globbing: globbingEnabledFromSettings(settingsCascade),
        streamSearch: aggregateStreamingSearch,
        batchChangesExecutionEnabled: isBatchChangesExecutionEnabled(settingsCascade),
        fetchHighlightedFileLineRanges: _fetchHighlightedFileLineRanges,
        batchChangesWebhookLogsEnabled: window.context.batchChangesWebhookLogsEnabled,
    } satisfies LegacyLayoutRouteComponentProps

    return <LegacyRouteContext.Provider value={legacyContext}>{children}</LegacyRouteContext.Provider>
}
export const LegacyRouteContext = createContext<LegacyLayoutRouteComponentProps | null>(null)
