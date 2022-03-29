import { isEqual } from 'lodash'
import { from, Observable, Unsubscribable } from 'rxjs'
import {
    distinctUntilChanged,
    filter,
    first,
    map,
    mapTo,
    publishReplay,
    refCount,
    switchMap,
    withLatestFrom,
} from 'rxjs/operators'
import type { HoverAlert } from 'sourcegraph'

import { ErrorLike, isErrorLike, isDefined, isNot } from '@sourcegraph/common'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { syncRemoteSubscription } from '@sourcegraph/shared/src/api/util'
import { Controller as ExtensionsController } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { MutationRecordLike } from '../../util/dom'

import { CodeHost } from './codeHost'
import { trackViews } from './views'

import styles from './nativeTooltips.module.scss'

const NATIVE_TOOLTIP_HIDDEN = styles.nativeTooltipHidden
const NATIVE_TOOLTIP_TYPE = 'nativeTooltips'

/**
 * Defines a native tooltip that is present on a page and exposes operations for manipulating it.
 */
export interface NativeTooltip {
    /** The native tooltip HTML element. */
    element: HTMLElement
}

export function handleNativeTooltips(
    mutations: Observable<MutationRecordLike[]>,
    nativeTooltipsEnabled: Observable<boolean>,
    { nativeTooltipResolvers, name }: Pick<CodeHost, 'nativeTooltipResolvers' | 'name' | 'getContext'>,
    repoSyncErrors: Observable<boolean>
): { nativeTooltipsAlert: Observable<HoverAlert>; subscription: Unsubscribable } {
    const nativeTooltips = mutations.pipe(trackViews(nativeTooltipResolvers || []))
    const nativeTooltipsAlert = nativeTooltips.pipe(
        first(),
        switchMap(() =>
            repoSyncErrors.pipe(
                filter(hasError => !hasError),
                mapTo({
                    type: NATIVE_TOOLTIP_TYPE,
                    summary: {
                        kind: MarkupKind.Markdown,
                        value: `<small>Sourcegraph has hidden ${name}'s native hover tooltips. You can toggle this at any time: to enable the native tooltips run "Code host: prefer non-Sourcegraph hover tooltips" from the command palette or set <code>{"codeHost.useNativeTooltips": true}</code> in your user settings.</small>`,
                    },
                })
            )
        ),
        publishReplay(1),
        refCount()
    )
    return {
        nativeTooltipsAlert,
        subscription: nativeTooltips.subscribe(({ element, subscriptions }) => {
            subscriptions.add(
                nativeTooltipsEnabled
                    .pipe(withLatestFrom(repoSyncErrors))
                    // This subscription is correctly handled through the view's `subscriptions`
                    // eslint-disable-next-line rxjs/no-nested-subscribe
                    .subscribe(([enabled, hasRepoSyncError]) => {
                        // If we can't provide the user hovers because it's private code, don't hide native tooltips.
                        // Otherwise we would have to show the user two alerts at the same time.
                        const isTooltipHidden = !enabled && !hasRepoSyncError
                        element.dataset.nativeTooltipHidden = String(isTooltipHidden)
                        element.classList.toggle(NATIVE_TOOLTIP_HIDDEN, isTooltipHidden)
                    })
            )
        }),
    }
}

export function nativeTooltipsEnabledFromSettings(settings: PlatformContext['settings']): Observable<boolean> {
    return from(settings).pipe(
        map(({ final }) => final),
        filter(isDefined),
        filter(isNot<ErrorLike | Settings, ErrorLike>(isErrorLike)),
        map(settings => !!settings['codeHost.useNativeTooltips']),
        distinctUntilChanged((a, b) => isEqual(a, b)),
        publishReplay(1),
        refCount()
    )
}

export function registerNativeTooltipContributions(
    extensionsController: Pick<ExtensionsController, 'extHostAPI'>
): Unsubscribable {
    return syncRemoteSubscription(
        extensionsController.extHostAPI.then(extensionHostAPI =>
            extensionHostAPI.registerContributions({
                actions: [
                    {
                        id: 'codeHost.toggleUseNativeTooltips',
                        command: 'updateConfiguration',
                        category: 'Code host',
                        commandArguments: [
                            'codeHost.useNativeTooltips',
                            /* eslint-disable-next-line no-template-curly-in-string */
                            '${!config.codeHost.useNativeTooltips}',
                            null,
                            'json',
                        ],
                        title:
                            /* eslint-disable-next-line no-template-curly-in-string */
                            'Prefer ${config.codeHost.useNativeTooltips && "Sourcegraph" || "non-Sourcegraph"} hover tooltips',
                    },
                ],
                menus: {
                    commandPalette: [
                        {
                            action: 'codeHost.toggleUseNativeTooltips',
                        },
                    ],
                },
            })
        )
    )
}
