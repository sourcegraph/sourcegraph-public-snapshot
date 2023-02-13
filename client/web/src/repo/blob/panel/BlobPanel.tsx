import React, { useEffect, useMemo } from 'react'

import { useLocation } from 'react-router-dom-v5-compat'
import { Observable, Subscription } from 'rxjs'

import { Panel, useBuiltinTabbedPanelViews } from '@sourcegraph/branded/src/components/panel/TabbedPanelContent'
import { PanelContent } from '@sourcegraph/branded/src/components/panel/views/PanelContent'
import { isDefined, isErrorLike } from '@sourcegraph/common'
import { FetchFileParameters } from '@sourcegraph/shared/src/backend/file'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings, SettingsCascadeOrError, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { AbsoluteRepoFile, ModeSpec, parseQueryAndHash } from '@sourcegraph/shared/src/util/url'

import { CodeIntelligenceProps } from '../../../codeintel'
import { ReferencesPanel } from '../../../codeintel/ReferencesPanel'
import { RepoRevisionSidebarCommits } from '../../RepoRevisionSidebarCommits'

interface Props
    extends AbsoluteRepoFile,
        ModeSpec,
        SettingsCascadeProps,
        ExtensionsControllerProps,
        ThemeProps,
        PlatformContextProps,
        Pick<CodeIntelligenceProps, 'useCodeIntel'>,
        TelemetryProps {
    repoID: Scalars['ID']
    repoName: string
    commitID: string

    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

export type BlobPanelTabID = 'info' | 'def' | 'references' | 'impl' | 'typedef' | 'history'

/**
 * A React hook that registers panel views for the blob.
 */
function useBlobPanelViews({
    extensionsController,
    revision,
    filePath,
    repoID,
    settingsCascade,
    isLightTheme,
    platformContext,
    useCodeIntel,
    telemetryService,
    fetchHighlightedFileLineRanges,
}: Props): void {
    const subscriptions = useMemo(() => new Subscription(), [])

    const preferAbsoluteTimestamps = preferAbsoluteTimestampsFromSettings(settingsCascade)
    const defaultPageSize = defaultPageSizeFromSettings(settingsCascade)

    const location = useLocation()

    const position = useMemo(() => {
        const parsedHash = parseQueryAndHash(location.search, location.hash)
        return parsedHash.line !== undefined
            ? { line: parsedHash.line, character: parsedHash.character || 0 }
            : undefined
    }, [location.hash, location.search])

    useBuiltinTabbedPanelViews(
        useMemo(() => {
            const panelDefinitions: Panel[] = [
                position
                    ? {
                          id: 'references',
                          title: 'References',
                          // The new reference panel contains definitions, references, and implementations. We need it to
                          // match all these IDs so it shows up when one of the IDs is used as `#tab=<ID>` in the URL.
                          matchesTabID: (id: string): boolean =>
                              id === 'def' || id === 'references' || id.startsWith('implementations_'),
                          element: (
                              <ReferencesPanel
                                  settingsCascade={settingsCascade}
                                  platformContext={platformContext}
                                  isLightTheme={isLightTheme}
                                  extensionsController={extensionsController}
                                  telemetryService={telemetryService}
                                  key="references"
                                  fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                                  useCodeIntel={useCodeIntel}
                              />
                          ),
                      }
                    : null,
                {
                    id: 'history',
                    title: 'History',
                    element: (
                        <PanelContent>
                            <RepoRevisionSidebarCommits
                                key="commits"
                                repoID={repoID}
                                revision={revision}
                                filePath={filePath}
                                preferAbsoluteTimestamps={preferAbsoluteTimestamps}
                                defaultPageSize={defaultPageSize}
                            />
                        </PanelContent>
                    ),
                },
            ].filter(isDefined)

            return panelDefinitions
        }, [
            position,
            settingsCascade,
            platformContext,
            isLightTheme,
            extensionsController,
            telemetryService,
            fetchHighlightedFileLineRanges,
            useCodeIntel,
            repoID,
            revision,
            filePath,
            preferAbsoluteTimestamps,
            defaultPageSize,
        ])
    )

    useEffect(() => () => subscriptions.unsubscribe(), [subscriptions])
}

function preferAbsoluteTimestampsFromSettings(settingsCascade: SettingsCascadeOrError<Settings>): boolean {
    if (settingsCascade.final && !isErrorLike(settingsCascade.final)) {
        return settingsCascade.final['history.preferAbsoluteTimestamps'] as boolean
    }
    return false
}

function defaultPageSizeFromSettings(settingsCascade: SettingsCascadeOrError<Settings>): number | undefined {
    if (settingsCascade.final && !isErrorLike(settingsCascade.final)) {
        return settingsCascade.final['history.defaultPageSize'] as number
    }

    return undefined
}

/**
 * Registers built-in tabbed panel views and renders `null`.
 */
export const BlobPanel: React.FunctionComponent<Props> = props => {
    useBlobPanelViews(props)

    return null
}
