import { HoverState } from '@sourcegraph/codeintellify'
import { getCodeElementsInRange, locateTarget } from '@sourcegraph/codeintellify/lib/token_position'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import * as H from 'history'
import { isEqual } from 'lodash'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { combineLatest, merge, ReplaySubject } from 'rxjs'
import { catchError, share, switchMap } from 'rxjs/operators'
import { ActionItemAction } from '../../../../shared/src/actions/ActionItem'
import { groupDecorationsByLine } from '../../../../shared/src/api/client/services/decoration'
import { HoverMerged } from '../../../../shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { HoverContext } from '../../../../shared/src/hover/HoverOverlay'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { asError, isErrorLike } from '../../../../shared/src/util/errors'
import {
    AbsoluteRepoFile,
    LineOrPositionOrRange,
    lprToSelectionsZeroIndexed,
    ModeSpec,
    parseHash,
    toPositionOrRangeHash,
    toURIWithPath,
} from '../../../../shared/src/util/url'
import { ThemeProps } from '../../../../shared/src/theme'
import { LineDecorator } from './LineDecorator'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { HoverThresholdProps } from '../RepoContainer'
import useDeepCompareEffect from 'use-deep-compare-effect'
import iterate from 'iterare'
import { Hoverifier } from '../../components/Hoverifier'
import { Falsy } from 'utility-types'

/**
 * toPortalID builds an ID that will be used for the {@link LineDecorator} portal containers.
 */
const toPortalID = (line: number): string => `line-decoration-attachment-${line}`

interface BlobProps
    extends SettingsCascadeProps,
        PlatformContextProps,
        TelemetryProps,
        HoverThresholdProps,
        ExtensionsControllerProps,
        ThemeProps {
    location: H.Location
    history: H.History
    className: string
    wrapCode: boolean
    /** The current text document to be rendered and provided to extensions */
    blobInfo: BlobInfo
}

export interface BlobInfo extends AbsoluteRepoFile, ThemeProps, ModeSpec {
    /** The raw content of the blob. */
    content: string

    /** The trusted syntax-highlighted code as HTML */
    html: string
}

const domFunctions = {
    getCodeElementFromTarget: (target: HTMLElement): HTMLTableCellElement | null => {
        // If the target is part of the line decoration attachment, return null.
        if (
            target.classList.contains('line-decoration-attachment') ||
            target.classList.contains('line-decoration-attachment__contents')
        ) {
            return null
        }

        const row = target.closest('tr')
        if (!row) {
            return null
        }
        return row.cells[1]
    },
    getCodeElementFromLineNumber: (codeView: HTMLElement, line: number): HTMLTableCellElement | null => {
        const table = codeView.firstElementChild as HTMLTableElement
        const row = table.rows[line - 1]
        if (!row) {
            return null
        }
        return row.cells[1]
    },
    getLineNumberFromCodeElement: (codeCell: HTMLElement): number => {
        const row = codeCell.closest('tr')
        if (!row) {
            throw new Error('Could not find closest row for codeCell')
        }
        const numberCell = row.cells[0]
        if (!numberCell || !numberCell.dataset.line) {
            throw new Error('Could not find line number')
        }
        return parseInt(numberCell.dataset.line, 10)
    },
}

export const Blob: React.FunctionComponent<BlobProps> = ({
    blobInfo,
    settingsCascade,
    telemetryService,
    isLightTheme,
    location,
    extensionsController,
    platformContext,
    history,
    className,
    wrapCode,
    onHoverShown,
}) => {
    // Emits on position changes from URL hash
    const locationPositions = useMemo(() => new ReplaySubject<LineOrPositionOrRange>(1), [])
    const nextLocationPosition = useCallback(
        (lineOrPositionOrRange: LineOrPositionOrRange) => locationPositions.next(lineOrPositionOrRange),
        [locationPositions]
    )
    const parsedHash = useMemo(() => parseHash(location.hash), [location.hash])
    useDeepCompareEffect(() => {
        // Update selected line when position in hash changes
        const codeView = codeViewReference.current
        if (codeView) {
            const codeCells = getCodeElementsInRange({
                codeView,
                position: parsedHash,
                getCodeElementFromLineNumber: domFunctions.getCodeElementFromLineNumber,
            })
            // Remove existing highlighting
            for (const selected of codeView.querySelectorAll('.selected')) {
                selected.classList.remove('selected')
            }
            for (const { element } of codeCells) {
                // Highlight row
                const row = element.parentElement as HTMLTableRowElement
                row.classList.add('selected')
            }
        }

        nextLocationPosition(parsedHash)
    }, [parsedHash])

    // Emits on blob info changes to update extension host model
    const blobInfoChanges = useMemo(() => new ReplaySubject<BlobInfo>(1), [])
    const nextBlobInfoChange = useCallback((blobInfo: BlobInfo) => blobInfoChanges.next(blobInfo), [blobInfoChanges])
    useEffect(() => {
        nextBlobInfoChange(blobInfo)
    }, [blobInfo, nextBlobInfoChange])

    // Update the Sourcegraph extensions model to reflect the current file.
    useEffect(() => {
        const subscription = combineLatest([blobInfoChanges, locationPositions]).subscribe(([blobInfo, position]) => {
            const uri = toURIWithPath(blobInfo)
            if (!extensionsController.services.model.hasModel(uri)) {
                extensionsController.services.model.addModel({
                    uri,
                    languageId: blobInfo.mode,
                    text: blobInfo.content,
                })
            }
            extensionsController.services.viewer.removeAllViewers()
            extensionsController.services.viewer.addViewer({
                type: 'CodeEditor' as const,
                resource: uri,
                selections: lprToSelectionsZeroIndexed(position),
                isActive: true,
            })
        })

        return () => {
            subscription.unsubscribe()
        }
    }, [extensionsController, blobInfoChanges, locationPositions])

    // When clicking a line, update the URL (which will in turn trigger a highlight of the line)
    const onLineSelection = useCallback(
        (event: MouseEvent, hoverState: HoverState<HoverContext, HoverMerged, ActionItemAction>): void => {
            // BUG: Sometimes requires two clicks to register click event
            const position = locateTarget(event.target as HTMLElement, domFunctions)
            let hash: string
            if (
                position &&
                event.shiftKey &&
                hoverState.selectedPosition &&
                hoverState.selectedPosition.line !== undefined
            ) {
                hash = toPositionOrRangeHash({
                    range: {
                        start: {
                            line: Math.min(hoverState.selectedPosition.line, position.line),
                        },
                        end: {
                            line: Math.max(hoverState.selectedPosition.line, position.line),
                        },
                    },
                })
            } else {
                hash = toPositionOrRangeHash({ position })
            }

            if (!hash.startsWith('#')) {
                hash = '#' + hash
            }

            history.push({ ...location, hash })
        },
        [history, location]
    )

    const singleClickGoToDefinition = useMemo(
        () =>
            Boolean(
                settingsCascade.final &&
                    !isErrorLike(settingsCascade.final) &&
                    settingsCascade.final.singleClickGoToDefinition === true
            ),
        [settingsCascade]
    )

    // Previously hovered token element and its event listener
    const hoveredTokenElement = useRef<{ element?: HTMLElement; eventListener?: (event: MouseEvent) => void }>({})
    const onHoverStateUpdate = useCallback(
        (hoverState: HoverState<HoverContext, HoverMerged, ActionItemAction>): void => {
            const { element, eventListener } = hoveredTokenElement.current
            if (singleClickGoToDefinition && element !== hoverState.hoveredTokenElement) {
                if (element) {
                    element.style.cursor = 'auto'
                    if (eventListener) {
                        element.removeEventListener('click', eventListener)
                    }
                    hoveredTokenElement.current = {}
                }

                if (hoverState.hoveredTokenElement) {
                    const goToDefinition = (event: MouseEvent): void => {
                        const goToDefinitionAction =
                            Array.isArray(hoverState.actionsOrError) &&
                            hoverState.actionsOrError.find(action => action.action.id === 'goToDefinition.preloaded')
                        if (goToDefinitionAction) {
                            history.push(goToDefinitionAction.action.commandArguments![0] as string)
                            event.stopPropagation()
                        }
                    }

                    hoverState.hoveredTokenElement.style.cursor = 'pointer'
                    hoverState.hoveredTokenElement.addEventListener('click', goToDefinition)
                    hoveredTokenElement.current = {
                        element: hoverState.hoveredTokenElement,
                        eventListener: goToDefinition,
                    }
                }
            }
        },
        [history, singleClickGoToDefinition]
    )

    const [decorationsOrError, setDecorationsOrError] = useState<TextDocumentDecoration[] | Error | null>()
    const codeViewReference = useRef<HTMLElement | null>()

    // Get decorations for the current file
    useEffect(() => {
        let lastBlobInfo: (AbsoluteRepoFile & ModeSpec) | undefined
        const decorations = blobInfoChanges.pipe(
            switchMap(blobInfo => {
                const blobInfoChanged = !isEqual(blobInfo, lastBlobInfo)
                lastBlobInfo = blobInfo // record so we can compute blobInfoChanged
                // Only clear decorations if the model changed. If only the extensions changed,
                // keep the old decorations until the new ones are available, to avoid UI jitter
                return merge(
                    blobInfoChanged ? [null] : [],
                    extensionsController.services.textDocumentDecoration.getDecorations({
                        uri: `git://${blobInfo.repoName}?${blobInfo.commitID}#${blobInfo.filePath}`,
                    })
                )
            }),
            share()
        )

        const subscription = decorations.pipe(catchError(error => [asError(error)])).subscribe(decorationsOrError => {
            setDecorationsOrError(decorationsOrError)
        })

        return () => {
            subscription.unsubscribe()
        }
    }, [blobInfoChanges, extensionsController])

    // Memoize `groupedDecorations` to avoid clearing and setting decorations in `LineDecorator`s on renders in which
    // decorations haven't changed.
    const groupedDecorations = useMemo(
        () =>
            decorationsOrError &&
            !isErrorLike(decorationsOrError) &&
            iterate(groupDecorationsByLine(decorationsOrError))
                .map(([line, decorations]) => {
                    const portalID = toPortalID(line)
                    return (
                        <LineDecorator
                            isLightTheme={isLightTheme}
                            key={`${portalID}-${blobInfo.filePath}`}
                            portalID={portalID}
                            getCodeElementFromLineNumber={domFunctions.getCodeElementFromLineNumber}
                            line={line}
                            decorations={decorations}
                            codeViewReference={codeViewReference}
                        />
                    )
                })
                .toArray(),

        [decorationsOrError, blobInfo.filePath, isLightTheme]
    )

    return (
        <Hoverifier<{
            blobInfo: BlobInfo
            groupedDecorations: JSX.Element[] | Falsy
            wrapCode: boolean
            className: string
        }>
            // Hover overlay props
            telemetryService={telemetryService}
            location={location}
            extensionsController={extensionsController}
            platformContext={platformContext}
            isLightTheme={isLightTheme}
            domFunctions={domFunctions}
            pinningEnabled={!singleClickGoToDefinition}
            absoluteRepoFile={blobInfo}
            // Observable that hoverifier depends on
            locationPositions={locationPositions}
            // Callbacks to hook into hoverifier state updates
            onHoverShown={onHoverShown}
            onHoverStateUpdate={onHoverStateUpdate}
            onLineSelection={onLineSelection}
            // Props to pass through
            passthroughProps={{ blobInfo, groupedDecorations, wrapCode, className }}
        >
            {useCallback(({ overlay, nextBlobElement, nextCodeViewElement, passthroughProps }) => {
                const { blobInfo, className, groupedDecorations, wrapCode } = passthroughProps

                return (
                    <div className={`blob ${className}`} ref={nextBlobElement}>
                        <code
                            className={`blob__code ${wrapCode ? ' blob__code--wrapped' : ''} test-blob`}
                            ref={codeView => {
                                // alternating between null and code view. https://github.com/facebook/react/issues/11258
                                console.log('ref', codeView)
                                codeViewReference.current = codeView
                                nextCodeViewElement(codeView)
                            }}
                            dangerouslySetInnerHTML={{ __html: blobInfo.html }}
                        />
                        {overlay}
                        {groupedDecorations}
                    </div>
                )
            }, [])}
        </Hoverifier>
    )
}
