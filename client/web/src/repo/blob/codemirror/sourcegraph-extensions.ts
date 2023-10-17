/**
 * This file contains CodeMirror extensions to integrate with Sourcegraph
 * extensions.
 *
 * This integration is done in various ways, see the specific extensions for
 * more information.
 */

import type { Extension } from '@codemirror/state'
import { type PluginValue, ViewPlugin, type ViewUpdate } from '@codemirror/view'
import type { Remote } from 'comlink'
import { combineLatest, EMPTY, from, type Observable, Subject, Subscription } from 'rxjs'
import { catchError, map, shareReplay, switchMap } from 'rxjs/operators'

import { type LineOrPositionOrRange, logger, lprToSelectionsZeroIndexed } from '@sourcegraph/common'
import type { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { haveInitialExtensionsLoaded } from '@sourcegraph/shared/src/api/features'
import type { ViewerId } from '@sourcegraph/shared/src/api/viewerTypes'
import type { RequiredExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { toURIWithPath } from '@sourcegraph/shared/src/util/url'

import type { BlobInfo } from '../CodeMirrorBlob'

import { type SelectedLineRange, selectedLines } from './linenumbers'

/**
 * Context holds all the information needed for CodeMirror extensions to
 * communicate with the extensions host.
 */
interface Context {
    viewerId: ViewerId
    extensionsController: RequiredExtensionsControllerProps['extensionsController']
    extensionHostAPI: Remote<FlatExtensionHostAPI>
    blobInfo: BlobInfo
}

/**
 * Enables integration with Sourcegraph extensions:
 * - Document highlights
 * - Hovercards (partially)
 * - Selection updates
 * - Reference panel warmup
 */
export function sourcegraphExtensions({
    blobInfo,
    initialSelection,
    extensionsController,
}: {
    blobInfo: BlobInfo
    initialSelection: LineOrPositionOrRange
    extensionsController: RequiredExtensionsControllerProps['extensionsController']
}): Extension {
    const subscriptions = new Subscription()

    // Initialize document and viewer as early as possible and make context
    // available as observable
    const contextObservable = from(
        extensionsController.extHostAPI.then(async extensionHostAPI => {
            const uri = toURIWithPath(blobInfo)

            const [, viewerId] = await Promise.all([
                // This call should be made before adding viewer, but since
                // messages to web worker are handled in order, we can use Promise.all
                extensionHostAPI.addTextDocumentIfNotExists({
                    uri,
                    languageId: blobInfo.mode,
                    text: blobInfo.content,
                }),
                extensionHostAPI.addViewerIfNotExists({
                    type: 'CodeEditor' as const,
                    resource: uri,
                    selections: lprToSelectionsZeroIndexed(initialSelection),
                    isActive: true,
                }),
            ])

            return [viewerId, extensionHostAPI] as [ViewerId, Remote<FlatExtensionHostAPI>]
        })
    ).pipe(
        catchError(() => {
            logger.error('Unable to initialize extensions context')
            return EMPTY
        }),
        map(([viewerId, extensionHostAPI]) => {
            subscriptions.add(() => {
                extensionHostAPI
                    .removeViewer(viewerId)
                    .catch(error => logger.error('Error removing viewer from extension host', error))
            })

            return {
                blobInfo,
                viewerId,
                extensionHostAPI,
                extensionsController,
            }
        }),
        shareReplay(1)
    )

    return [
        // This view plugin is used to have a way to cleanup any resources via the
        // `destroy` method.
        ViewPlugin.define(() => ({
            destroy() {
                subscriptions.unsubscribe()
            },
        })),
        ViewPlugin.define(() => new SelectionManager(contextObservable)),
        ViewPlugin.define(() => new WarmupReferencesManager(contextObservable)),
    ]
}

//
// Selection change notifier
//

/**
 * The selection manager listens to CodeMirror selection changes and sends them
 * to the extensions host.
 */
class SelectionManager implements PluginValue {
    private nextSelection: Subject<SelectedLineRange> = new Subject()
    private subscription = new Subscription()

    constructor(context: Observable<Context>) {
        this.subscription = combineLatest([context, this.nextSelection]).subscribe(([context, selection]) => {
            // Used to convert SelectedLineRange type to a valid LineOrPositionOrRange type to keep TypeScript happy
            let lprSelection: LineOrPositionOrRange = {}
            if (selection) {
                lprSelection =
                    selection.line !== selection.endLine
                        ? { line: selection.line, endLine: selection.endLine }
                        : { line: selection.line, character: selection.character }
            }
            context.extensionHostAPI
                .setEditorSelections(context.viewerId, lprToSelectionsZeroIndexed(lprSelection))
                .catch(error => logger.error('Error updating editor selections on extension host', error))
        })
    }

    public destroy(): void {
        this.subscription.unsubscribe()
    }

    public update(update: ViewUpdate): void {
        if (update.state.field(selectedLines) !== update.startState.field(selectedLines)) {
            this.nextSelection.next(update.state.field(selectedLines))
        }
    }
}

class WarmupReferencesManager implements PluginValue {
    private subscription: Subscription

    constructor(context: Observable<Context>) {
        this.subscription = context
            .pipe(switchMap(context => haveInitialExtensionsLoaded(context.extensionsController.extHostAPI)))
            .subscribe()
    }

    public destroy(): void {
        this.subscription.unsubscribe()
    }
}
