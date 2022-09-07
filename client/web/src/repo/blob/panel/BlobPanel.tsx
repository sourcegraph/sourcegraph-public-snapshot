import React, { useEffect, useMemo, useRef } from 'react'

import * as H from 'history'
import { EMPTY, from, ReplaySubject, Subscription } from 'rxjs'
import { map, mapTo, switchMap, tap } from 'rxjs/operators'

import {
    BuiltinTabbedPanelDefinition,
    useBuiltinTabbedPanelViews,
} from '@sourcegraph/branded/src/components/panel/TabbedPanelContent'
import { TextDocumentPositionParameters } from '@sourcegraph/client-api'
import { isErrorLike } from '@sourcegraph/common'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { Activation, ActivationProps } from '@sourcegraph/shared/src/components/activation/Activation'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings, SettingsCascadeOrError, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { AbsoluteRepoFile, ModeSpec, parseQueryAndHash, UIPositionSpec } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/wildcard'

import { ReferencesPanelWithMemoryRouter } from '../../../codeintel/ReferencesPanel'
import { RepoRevisionSidebarCommits } from '../../RepoRevisionSidebarCommits'

interface Props
    extends AbsoluteRepoFile,
        ModeSpec,
        SettingsCascadeProps,
        ExtensionsControllerProps,
        ActivationProps,
        ThemeProps,
        PlatformContextProps,
        TelemetryProps {
    location: H.Location
    history: H.History
    repoID: Scalars['ID']
    repoName: string
    commitID: string
}

export type BlobPanelTabID = 'info' | 'def' | 'references' | 'impl' | 'typedef' | 'history'

/** The subject (what the contextual information refers to). */
interface PanelSubject extends AbsoluteRepoFile, ModeSpec, Partial<UIPositionSpec> {
    repoID: string

    /**
     * Include the full URI fragment here because it represents the state of panels, and we want
     * panels to be re-rendered when this state changes.
     */
    hash: string

    history: H.History
    location: H.Location
}

/**
 * A React hook that registers panel views for the blob.
 */
function useBlobPanelViews({
    extensionsController,
    activation,
    repoName,
    commitID,
    revision,
    mode,
    filePath,
    repoID,
    location,
    history,
    settingsCascade,
    isLightTheme,
    platformContext,
    telemetryService,
}: Props): void {
    const subscriptions = useMemo(() => new Subscription(), [])

    // Activation props are not stable
    const activationReference = useRef<Activation | undefined>(activation)
    activationReference.current = activation

    // Keep active code editor position subscription active to prevent empty loading state
    // (for main thread -> ext host -> main thread roundtrip for editor position)
    const activeCodeEditorPositions = useMemo(() => new ReplaySubject<TextDocumentPositionParameters | null>(1), [])
    useObservable(
        useMemo(() => {
            if (extensionsController === null) {
                return EMPTY
            }
            return from(extensionsController.extHostAPI).pipe(
                switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getActiveCodeEditorPosition())),
                tap(parameters => activeCodeEditorPositions.next(parameters)),
                mapTo(undefined)
            )
        }, [activeCodeEditorPositions, extensionsController])
    )

    const preferAbsoluteTimestamps = preferAbsoluteTimestampsFromSettings(settingsCascade)

    // Source for history panel
    const panelSubject = useMemo(() => {
        const parsedHash = parseQueryAndHash(location.search, location.hash)
        return {
            repoID,
            repoName,
            commitID,
            revision,
            filePath,
            mode,
            position:
                parsedHash.line !== undefined
                    ? { line: parsedHash.line, character: parsedHash.character || 0 }
                    : undefined,
            hash: location.hash,
            history,
            location,
        }
    }, [commitID, filePath, history, location, mode, repoID, repoName, revision])

    const panelSubjectChanges = useMemo(() => new ReplaySubject<PanelSubject>(1), [])
    useEffect(() => {
        panelSubjectChanges.next(panelSubject)
    }, [panelSubject, panelSubjectChanges])

    useBuiltinTabbedPanelViews(
        useMemo(() => {
            const panelDefinitions: BuiltinTabbedPanelDefinition[] = [
                {
                    id: 'history',
                    provider: panelSubjectChanges.pipe(
                        map(({ repoID, revision, filePath, history, location }) => ({
                            title: 'History',
                            content: '',
                            priority: 150,
                            selector: null,
                            locationProvider: undefined,
                            reactElement: (
                                <RepoRevisionSidebarCommits
                                    key="commits"
                                    repoID={repoID}
                                    revision={revision}
                                    filePath={filePath}
                                    history={history}
                                    location={location}
                                    preferAbsoluteTimestamps={preferAbsoluteTimestamps}
                                />
                            ),
                        }))
                    ),
                },
            ]

            panelDefinitions.push({
                id: 'references',
                provider: panelSubjectChanges.pipe(
                    map(({ position, history, location }) => ({
                        title: 'References',
                        content: '',
                        priority: 180,
                        selector: null,
                        locationProvider: undefined,
                        // The new reference panel contains definitoins, references, and implementations. We need it to
                        // match all these IDs so it shows up when one of the IDs is used as `#tab=<ID>` in the URL.
                        matchesTabID: (id: string): boolean =>
                            id === 'def' || id === 'references' || id.startsWith('implementations_'),
                        // This panel doesn't need a wrapper
                        noWrapper: true,
                        reactElement: position ? (
                            <ReferencesPanelWithMemoryRouter
                                settingsCascade={settingsCascade}
                                platformContext={platformContext}
                                isLightTheme={isLightTheme}
                                extensionsController={extensionsController}
                                telemetryService={telemetryService}
                                key="references"
                                externalHistory={history}
                                externalLocation={location}
                            />
                        ) : (
                            <></>
                        ),
                    }))
                ),
            })

            return panelDefinitions
        }, [
            panelSubjectChanges,
            preferAbsoluteTimestamps,
            isLightTheme,
            settingsCascade,
            telemetryService,
            platformContext,
            extensionsController,
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

/**
 * Registers built-in tabbed panel views and renders `null`.
 */
export const BlobPanel: React.FunctionComponent<Props> = props => {
    useBlobPanelViews(props)

    return null
}
