import { isEqual } from 'lodash'
import { from, Observable, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, filter, map, publishReplay, refCount, tap } from 'rxjs/operators'
import { parseTemplate } from '../../../../shared/src/api/client/context/expr/evaluator'
import { Services } from '../../../../shared/src/api/client/services'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { Settings } from '../../../../shared/src/settings/settings'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { isDefined } from '../../../../shared/src/util/types'
import { MutationRecordLike } from '../../shared/util/dom'
import { CodeHost } from './code_intelligence'
import { trackViews } from './views'

const NATIVE_TOOLTIP_HIDDEN = 'native-tooltip--hidden'

/**
 * Defines a native tooltip that is present on a page and exposes operations for manipulating it.
 */
export interface NativeTooltip {
    /** The native tooltip HTML element. */
    element: HTMLElement
}

/**
 * Handles added and removed native tooltips according to the {@link CodeHost} configuration.
 */
export function handleNativeTooltips(
    mutations: Observable<MutationRecordLike[]>,
    nativeTooltipsEnabled: Observable<boolean>,
    nativeTooltipResolvers: NonNullable<CodeHost['nativeTooltipResolvers']>
): Unsubscribable {
    /** A stream of added or removed native tooltips. */
    const nativeTooltips = mutations.pipe(trackViews(nativeTooltipResolvers))

    return nativeTooltips.subscribe(({ element, subscriptions }) => {
        subscriptions.add(
            nativeTooltipsEnabled
                .pipe(
                    tap(enabled => {
                        if (enabled) {
                            element.classList.remove(NATIVE_TOOLTIP_HIDDEN)
                        } else {
                            element.classList.add(NATIVE_TOOLTIP_HIDDEN)
                        }
                    })
                )
                // This subscription is correctly handled through the view's subscriptions.
                .subscribe()
        )
    })
}

export function nativeTooltipsEnabledFromSettings(settings: PlatformContext['settings']): Observable<boolean> {
    return from(settings).pipe(
        map(({ final }) => final),
        filter(isDefined),
        filter((s: Settings | ErrorLike): s is Settings => !isErrorLike(s)),
        map(s => !!s['codeHost.useNativeTooltips']),
        distinctUntilChanged(isEqual),
        publishReplay(1),
        refCount()
    )
}

export function registerNativeTooltipContributions(extensionsController: {
    services: Pick<Services, 'contribution'>
}): Unsubscribable {
    return extensionsController.services.contribution.registerContributions({
        contributions: {
            actions: [
                {
                    id: 'codeHost.toggleUseNativeTooltips',
                    command: 'updateConfiguration',
                    category: parseTemplate('Code host'),
                    commandArguments: [
                        parseTemplate('codeHost.useNativeTooltips'),
                        parseTemplate('${!config.codeHost.useNativeTooltips}'),
                        null,
                        parseTemplate('json'),
                    ],
                    title: parseTemplate(
                        'Use ${config.codeHost.useNativeTooltips && "Sourcegraph" || "native"} hover tooltips'
                    ),
                },
            ],
            menus: {
                commandPalette: [
                    {
                        action: 'codeHost.toggleUseNativeTooltips',
                    },
                ],
            },
        },
    })
}
