import { uniqueId } from 'lodash'
import { of, Subject, Subscription } from 'rxjs'
import { first } from 'rxjs/operators'
import { MarkupKind } from 'sourcegraph'
import { LinkPreviewMerged } from '../../../../../shared/src/api/client/services/linkPreview'
import { createBarrier } from '../../../../../shared/src/api/integration-test/testHelpers'
import { handleContentViews } from './content_views'

describe('content_views', () => {
    beforeEach(() => {
        document.body.innerHTML = ''
    })

    describe('handleContentViews()', () => {
        let subscriptions = new Subscription()

        afterEach(() => {
            subscriptions.unsubscribe()
            subscriptions = new Subscription()
        })

        const createTestElement = () => {
            const el = document.createElement('div')
            el.className = `test test-${uniqueId()}`
            document.body.appendChild(el)
            return el
        }

        test('detects and annotates content views', async () => {
            const element = createTestElement()
            element.id = 'content-view'
            element.innerHTML = '0 <a href=#foo>foo</a> 1 <a href=#bar>bar</a> 2 <a href=#qux>qux</a> 3'
            const wait = new Subject<void>()
            subscriptions.add(
                handleContentViews(
                    of([{ addedNodes: [document.body], removedNodes: [] }]),
                    {
                        extensionsController: {
                            services: {
                                linkPreviews: {
                                    provideLinkPreview: url => {
                                        wait.next()
                                        if (url.includes('bar')) {
                                            return of(null)
                                        }
                                        return of<LinkPreviewMerged>({
                                            content: [
                                                {
                                                    kind: 'markdown' as MarkupKind.Markdown,
                                                    value: `**${url.slice(url.lastIndexOf('#') + 1)}** x`,
                                                },
                                            ],
                                            hover: [
                                                {
                                                    kind: 'plaintext' as MarkupKind.PlainText,
                                                    value: url.slice(url.lastIndexOf('#') + 1),
                                                },
                                            ],
                                        })
                                    },
                                },
                            },
                        },
                    },
                    {
                        contentViewResolvers: [{ selector: 'div', resolveView: () => ({ element }) }],
                        setElementTooltip: (e, text) =>
                            text !== null ? e.setAttribute('data-tooltip', text) : e.removeAttribute('data-tooltip'),
                    }
                )
            )
            await wait.pipe(first()).toPromise()
            expect(element.classList.contains('sg-mounted')).toBe(true)
            expect(element.innerHTML).toBe(
                '0 <a href="#foo" data-tooltip="foo">foo</a><span class="sg-link-preview-content" data-tooltip="foo"><strong>foo</strong> x</span> 1 <a href="#bar">bar</a> 2 <a href="#qux" data-tooltip="qux">qux</a><span class="sg-link-preview-content" data-tooltip="qux"><strong>qux</strong> x</span> 3'
            )

            // Mutate content view.
            element.innerHTML = '4 <a href=#zip>zip</a> 5'
            await wait.pipe(first()).toPromise()
            expect(element.classList.contains('sg-mounted')).toBe(true)
            expect(element.innerHTML).toBe(
                '4 <a href="#zip" data-tooltip="zip">zip</a><span class="sg-link-preview-content" data-tooltip="zip"><strong>zip</strong> x</span> 5'
            )
        })

        test('handles multiple emissions', async () => {
            const element = createTestElement()
            element.id = 'content-view'
            element.innerHTML = '0 <a href=#foo>foo</a> 1 <a href=#bar>bar</a> 2'
            const originalInnerHTML = element.innerHTML
            const fooLinkPreviewValues = new Subject<LinkPreviewMerged>()
            const { wait, done } = createBarrier()
            subscriptions.add(
                handleContentViews(
                    of([{ addedNodes: [document.body], removedNodes: [] }]),
                    {
                        extensionsController: {
                            services: {
                                linkPreviews: {
                                    provideLinkPreview: url => {
                                        done()
                                        return url.includes('bar') ? of(null) : fooLinkPreviewValues
                                    },
                                },
                            },
                        },
                    },
                    {
                        contentViewResolvers: [{ selector: 'div', resolveView: () => ({ element }) }],
                        setElementTooltip: (e, text) =>
                            text !== null ? e.setAttribute('data-tooltip', text) : e.removeAttribute('data-tooltip'),
                    }
                )
            )

            await wait
            expect(element.classList.contains('sg-mounted')).toBe(true)
            expect(element.innerHTML).toBe(originalInnerHTML)

            fooLinkPreviewValues.next({
                content: [
                    {
                        kind: 'markdown' as MarkupKind.Markdown,
                        value: `**foo**`,
                    },
                ],
                hover: [
                    {
                        kind: 'plaintext' as MarkupKind.PlainText,
                        value: 'foo',
                    },
                ],
            })
            expect(element.innerHTML).toBe(
                '0 <a href="#foo" data-tooltip="foo">foo</a><span class="sg-link-preview-content" data-tooltip="foo"><strong>foo</strong></span> 1 <a href="#bar">bar</a> 2'
            )

            fooLinkPreviewValues.next({
                content: [
                    {
                        kind: 'markdown' as MarkupKind.Markdown,
                        value: `**foo2**`,
                    },
                ],
                hover: [
                    {
                        kind: 'plaintext' as MarkupKind.PlainText,
                        value: 'foo2',
                    },
                ],
            })
            expect(element.innerHTML).toBe(
                '0 <a href="#foo" data-tooltip="foo2">foo</a><span class="sg-link-preview-content" data-tooltip="foo2"><strong>foo2</strong></span> 1 <a href="#bar">bar</a> 2'
            )

            fooLinkPreviewValues.next({
                content: [],
                hover: [
                    {
                        kind: 'plaintext' as MarkupKind.PlainText,
                        value: 'foo2',
                    },
                ],
            })
            expect(element.innerHTML).toBe('0 <a href="#foo" data-tooltip="foo2">foo</a> 1 <a href="#bar">bar</a> 2')

            fooLinkPreviewValues.next({
                content: [],
                hover: [],
            })
            expect(element.innerHTML).toBe('0 <a href="#foo">foo</a> 1 <a href="#bar">bar</a> 2')
        })
    })
})
