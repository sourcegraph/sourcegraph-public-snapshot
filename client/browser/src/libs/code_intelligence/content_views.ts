import { animationFrameScheduler, merge, Observable, of, Subject, Subscription, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, map, mapTo, mergeMap, observeOn, tap, throttleTime } from 'rxjs/operators'
import { LinkPreviewProviderRegistry } from '../../../../../shared/src/api/client/services/linkPreview'
import { applyLinkPreview } from '../../../../../shared/src/components/linkPreviews/linkPreviews'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { MutationRecordLike, observeMutations } from '../../shared/util/dom'
import { CodeHost } from './code_intelligence'
import { trackViews } from './views'

/**
 * Defines a content view that is present on a page and exposes operations for manipulating it.
 */
export interface ContentView {
    /** The content view HTML element. */
    element: HTMLElement
}

/**
 * Handles added and removed content views according to the {@link CodeHost} configuration.
 */
export function handleContentViews(
    mutations: Observable<MutationRecordLike[]>,
    {
        extensionsController,
    }:
        | ExtensionsControllerProps
        | {
              extensionsController: {
                  services: { linkPreviews: Pick<LinkPreviewProviderRegistry, 'provideLinkPreview'> }
              }
          },
    {
        contentViewResolvers,
        linkPreviewContentClass,
        setElementTooltip,
    }: Pick<CodeHost, 'contentViewResolvers' | 'linkPreviewContentClass' | 'setElementTooltip'>
): Unsubscribable {
    /** A stream of added or removed content views. */
    const contentViews = mutations.pipe(
        trackViews<ContentView>(contentViewResolvers || []),
        observeOn(animationFrameScheduler)
    )
    interface ContentViewState {
        subscriptions: Subscription
    }
    /** Map from content view element to the state associated with it (to be updated or removed). */
    const contentViewStates = new Map<Element, ContentViewState>()

    /** Pause DOM MutationObserver while we are making changes to avoid duplicating work. */
    const pauseMutationObserver = new Subject<boolean>()

    return contentViews
        .pipe(
            mergeMap(contentViewEvent => {
                if (contentViewEvent.type === 'added') {
                    return merge(
                        of(contentViewEvent),

                        /**
                         * Observe updates to the element. Only emit on mutations that actually
                         * change the innerHTML so that our own {@link applyLinkPreview} updates
                         * don't trigger needless work. It is not sufficient to suppress observing
                         * these changes using {@link MutationObserver#disconnect} because that does
                         * not actually seem to suppress mutation notifications in tests when using
                         * jsdom.
                         */
                        observeMutations(contentViewEvent.element, { childList: true }, pauseMutationObserver).pipe(
                            observeOn(animationFrameScheduler),
                            map(() => contentViewEvent.element.innerHTML),
                            distinctUntilChanged(),
                            mapTo({ type: 'updated' as const, element: contentViewEvent.element }),
                            throttleTime(2000, undefined, { leading: true, trailing: true }) // reduce the harm from an infinite loop bug
                        )
                    )
                }
                return of(contentViewEvent)
            }),
            tap(contentViewEvent => console.log(`Content view ${contentViewEvent.type}`, { contentViewEvent })),
            tap(contentViewEvent => {
                // Handle added, updated, or removed content views.

                if (contentViewEvent.type === 'removed' || contentViewEvent.type === 'updated') {
                    const contentViewState = contentViewStates.get(contentViewEvent.element)
                    if (contentViewState) {
                        contentViewState.subscriptions.unsubscribe()
                        contentViewStates.delete(contentViewEvent.element)
                    }
                }

                if (contentViewEvent.type === 'added' || contentViewEvent.type === 'updated') {
                    const { element } = contentViewEvent
                    let contentViewState = contentViewStates.get(contentViewEvent.element)
                    if (!contentViewState) {
                        contentViewState = { subscriptions: new Subscription() }
                    }

                    // Add link preview content.
                    for (const link of element.querySelectorAll<HTMLAnchorElement>('a[href]')) {
                        contentViewState.subscriptions.add(
                            extensionsController.services.linkPreviews
                                .provideLinkPreview(link.href)
                                // The nested subscribe cannot be replaced with a switchMap()
                                // because we are managing a stateful Map. The subscription is
                                // managed correctly.
                                //
                                // tslint:disable-next-line: rxjs-no-nested-subscribe
                                .subscribe(linkPreview => {
                                    try {
                                        pauseMutationObserver.next(true) // ignore DOM mutations we make
                                        applyLinkPreview(
                                            { setElementTooltip, linkPreviewContentClass },
                                            link,
                                            linkPreview
                                        )
                                    } finally {
                                        pauseMutationObserver.next(false) // stop ignoring DOM mutations
                                    }
                                })
                        )
                    }
                }

                if (contentViewEvent.type === 'added') {
                    contentViewEvent.element.classList.add('sg-mounted')
                }
            })
        )
        .subscribe()
}
