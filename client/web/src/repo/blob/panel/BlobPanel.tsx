import { useCallback, useEffect, useMemo, useRef } from 'react'

import * as H from 'history'
import { from, Observable, ReplaySubject, Subscription } from 'rxjs'
import { map, mapTo, switchMap, tap } from 'rxjs/operators'

import {
    BuiltinTabbedPanelDefinition,
    BuiltinTabbedPanelView,
    useBuiltinTabbedPanelViews,
} from '@sourcegraph/branded/src/components/panel/TabbedPanelContent'
import { ReferenceParameters, TextDocumentPositionParameters } from '@sourcegraph/client-api'
import { MaybeLoadingResult } from '@sourcegraph/codeintellify'
import { isErrorLike } from '@sourcegraph/common'
import * as clientType from '@sourcegraph/extension-api-types'
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
export function useBlobPanelViews({
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
        useMemo(
            () =>
                from(extensionsController.extHostAPI).pipe(
                    switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getActiveCodeEditorPosition())),
                    tap(parameters => activeCodeEditorPositions.next(parameters)),
                    mapTo(undefined)
                ),
            [activeCodeEditorPositions, extensionsController]
        )
    )

    const maxPanelResults = maxPanelResultsFromSettings(settingsCascade)
    const preferAbsoluteTimestamps = preferAbsoluteTimestampsFromSettings(settingsCascade)
    const experimentalReferencePanelEnabled =
        !isErrorLike(settingsCascade.final) && settingsCascade.final?.experimentalFeatures?.coolCodeIntel === true

    // Creates source for definition and reference panels
    const createLocationProvider = useCallback(
        <P extends TextDocumentPositionParameters>(
            id: string,
            title: string,
            priority: number,
            provideLocations: (parameters: P) => Observable<MaybeLoadingResult<clientType.Location[]>>,
            extraParameters?: Pick<P, Exclude<keyof P, keyof TextDocumentPositionParameters>>
        ): Observable<BuiltinTabbedPanelView | null> =>
            activeCodeEditorPositions.pipe(
                map(textDocumentPositionParameters => {
                    if (!textDocumentPositionParameters) {
                        return null
                    }

                    return {
                        title,
                        content: '',
                        selector: null,
                        priority,

                        maxLocationResults: id === 'references' || id === 'def' ? maxPanelResults : undefined,
                        // This disable directive is necessary because TypeScript is not yet smart
                        // enough to know that (typeof params & typeof extraParams) is P.
                        //
                        // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                        locationProvider: provideLocations({
                            ...textDocumentPositionParameters,
                            ...extraParameters,
                        } as P).pipe(
                            tap(({ result: locations }) => {
                                if (activationReference.current && id === 'references' && locations.length > 0) {
                                    activationReference.current.update({ FoundReferences: true })
                                }
                            })
                        ),
                    }
                })
            ),
        [activeCodeEditorPositions, maxPanelResults]
    )

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

            if (!experimentalReferencePanelEnabled) {
                panelDefinitions.push(
                    ...[
                        {
                            id: 'def',
                            provider: createLocationProvider('def', 'Definition', 190, parameters =>
                                from(extensionsController.extHostAPI).pipe(
                                    switchMap(extensionHostAPI =>
                                        wrapRemoteObservable(extensionHostAPI.getDefinition(parameters))
                                    )
                                )
                            ),
                        },
                        {
                            id: 'references',
                            provider: createLocationProvider<ReferenceParameters>(
                                'references',
                                'References',
                                180,
                                parameters =>
                                    from(extensionsController.extHostAPI).pipe(
                                        switchMap(extensionHostAPI =>
                                            wrapRemoteObservable(
                                                extensionHostAPI.getReferences(parameters, {
                                                    includeDeclaration: false,
                                                })
                                            )
                                        )
                                    )
                            ),
                        },
                    ]
                )
            } else {
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
            }

            return panelDefinitions
        }, [
            experimentalReferencePanelEnabled,
            createLocationProvider,
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

function maxPanelResultsFromSettings(settingsCascade: SettingsCascadeOrError<Settings>): number | undefined {
    if (settingsCascade.final && !isErrorLike(settingsCascade.final)) {
        return settingsCascade.final['codeIntelligence.maxPanelResults'] as number
    }
    return undefined
}

function preferAbsoluteTimestampsFromSettings(settingsCascade: SettingsCascadeOrError<Settings>): boolean {
    if (settingsCascade.final && !isErrorLike(settingsCascade.final)) {
        return settingsCascade.final['history.preferAbsoluteTimestamps'] as boolean
    }
    return false
}
