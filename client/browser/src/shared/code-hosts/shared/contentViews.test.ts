import { nextTick } from 'process'
import { promisify } from 'util'

import { uniqueId } from 'lodash'
import { concat, Observable, of, ReplaySubject, Subject, Subscription } from 'rxjs'
import { first, tap } from 'rxjs/operators'

import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { LinkPreviewMerged } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { pretendProxySubscribable, pretendRemote } from '@sourcegraph/shared/src/api/util'

import { MutationRecordLike } from '../../util/dom'

import { handleContentViews } from './contentViews'

const tick = promisify(nextTick)

describe('contentViews', () => {
    beforeEach(() => {
        document.body.innerHTML = ''
    })

    describe('handleContentViews()', () => {
        let subscriptions = new Subscription()

        afterEach(() => {
            subscriptions.unsubscribe()
            subscriptions = new Subscription()
        })

        const createTestElement = (): HTMLElement => {
            const element = document.createElement('div')
            element.className = `test test-${uniqueId()}`
            document.body.append(element)
            return element
        }

        test('detects addition, mutation, and removal of content views (and annotates them)', async () => {
            const element = createTestElement()
            element.id = 'content-view'
            element.innerHTML = '0 <a href=#foo>foo</a> 1 <a href=#bar>bar</a> 2 <a href=#qux>qux</a> 3'

            const waitSubject = new Subject<void>()
            const unsubscribed = new Subject<void>()
            const mutations = new Subject<MutationRecordLike[]>()

            subscriptions.add(
                handleContentViews(
                    mutations,
                    {
                        extensionsController: {
                            extHostAPI: Promise.resolve(
                                pretendRemote<FlatExtensionHostAPI>({
                                    getLinkPreviews: url => {
                                        waitSubject.next()

                                        if (url.includes('bar')) {
                                            return pretendProxySubscribable(of(null))
                                        }

                                        return pretendProxySubscribable(
                                            concat(
                                                of<LinkPreviewMerged>({
                                                    content: [
                                                        {
                                                            kind: MarkupKind.Markdown,
                                                            value: `**${url.slice(url.lastIndexOf('#') + 1)}** x`,
                                                        },
                                                    ],
                                                    hover: [
                                                        {
                                                            kind: MarkupKind.PlainText,
                                                            value: url.slice(url.lastIndexOf('#') + 1),
                                                        },
                                                    ],
                                                }),
                                                // Support checking that the provider's observable was unsubscribed.
                                                new Observable<LinkPreviewMerged>(() => () => {
                                                    unsubscribed.next()
                                                })
                                            )
                                        )
                                    },
                                })
                            ),
                        },
                    },
                    {
                        contentViewResolvers: [{ selector: 'div', resolveView: () => ({ element }) }],
                        setElementTooltip: (element, text) =>
                            text !== null ? (element.dataset.tooltip = text) : element.removeAttribute('data-tooltip'),
                    }
                )
            )
            // Add content view.
            mutations.next([{ addedNodes: [document.body], removedNodes: [] }])
            await waitSubject.pipe(first()).toPromise()
            await tick()
            expect(element.innerHTML).toBe(
                '0 <a href="#foo" data-tooltip="foo">foo</a><span class="sg-link-preview-content" data-tooltip="foo"><strong>foo</strong> x</span> 1 <a href="#bar">bar</a> 2 <a href="#qux" data-tooltip="qux">qux</a><span class="sg-link-preview-content" data-tooltip="qux"><strong>qux</strong> x</span> 3'
            )

            // Mutate content view.
            element.innerHTML = '4 <a href=#zip>zip</a> 5'
            await Promise.all([unsubscribed.pipe(first()).toPromise(), waitSubject.pipe(first()).toPromise()])
            await tick()

            expect(element.innerHTML).toBe(
                '4 <a href="#zip" data-tooltip="zip">zip</a><span class="sg-link-preview-content" data-tooltip="zip"><strong>zip</strong> x</span> 5'
            )

            // Remove content view.
            mutations.next([{ addedNodes: [], removedNodes: [element] }])
            await unsubscribed.pipe(first()).toPromise()
            await tick()
        })

        test('handles multiple emissions', async () => {
            const element = createTestElement()
            element.id = 'content-view'
            element.innerHTML = '0 <a href=#foo>foo</a> 1 <a href=#bar>bar</a> 2'
            const originalInnerHTML = element.innerHTML
            const fooLinkPreviewValues = new ReplaySubject<LinkPreviewMerged>(1)
            const waitSubject = new Subject<void>()
            subscriptions.add(
                handleContentViews(
                    of([{ addedNodes: [document.body], removedNodes: [] }]),
                    {
                        extensionsController: {
                            extHostAPI: Promise.resolve(
                                pretendRemote<FlatExtensionHostAPI>({
                                    getLinkPreviews: url => {
                                        waitSubject.next()
                                        return pretendProxySubscribable(
                                            url.includes('bar')
                                                ? of(null)
                                                : fooLinkPreviewValues.pipe(tap(() => waitSubject.next()))
                                        )
                                    },
                                })
                            ),
                        },
                    },
                    {
                        contentViewResolvers: [{ selector: 'div', resolveView: () => ({ element }) }],
                        setElementTooltip: (element, text) =>
                            text !== null ? (element.dataset.tooltip = text) : element.removeAttribute('data-tooltip'),
                    }
                )
            )

            let wait: Promise<void> = waitSubject.pipe(first()).toPromise()
            await wait
            expect(element.innerHTML).toBe(originalInnerHTML)

            wait = waitSubject.pipe(first()).toPromise()
            fooLinkPreviewValues.next({
                content: [
                    {
                        kind: MarkupKind.Markdown,
                        value: '**foo**',
                    },
                ],
                hover: [
                    {
                        kind: MarkupKind.PlainText,
                        value: 'foo',
                    },
                ],
            })
            await wait
            expect(element.innerHTML).toBe(
                '0 <a href="#foo" data-tooltip="foo">foo</a><span class="sg-link-preview-content" data-tooltip="foo"><strong>foo</strong></span> 1 <a href="#bar">bar</a> 2'
            )

            wait = waitSubject.pipe(first()).toPromise()
            fooLinkPreviewValues.next({
                content: [
                    {
                        kind: MarkupKind.Markdown,
                        value: '**foo2**',
                    },
                ],
                hover: [
                    {
                        kind: MarkupKind.PlainText,
                        value: 'foo2',
                    },
                ],
            })
            await wait
            expect(element.innerHTML).toBe(
                '0 <a href="#foo" data-tooltip="foo2">foo</a><span class="sg-link-preview-content" data-tooltip="foo2"><strong>foo2</strong></span> 1 <a href="#bar">bar</a> 2'
            )

            wait = waitSubject.pipe(first()).toPromise()
            fooLinkPreviewValues.next({
                content: [],
                hover: [
                    {
                        kind: MarkupKind.PlainText,
                        value: 'foo2',
                    },
                ],
            })
            await wait
            expect(element.innerHTML).toBe('0 <a href="#foo" data-tooltip="foo2">foo</a> 1 <a href="#bar">bar</a> 2')

            wait = waitSubject.pipe(first()).toPromise()
            fooLinkPreviewValues.next({
                content: [],
                hover: [],
            })
            await wait
            expect(element.innerHTML).toBe('0 <a href="#foo">foo</a> 1 <a href="#bar">bar</a> 2')
        })
    })
})
