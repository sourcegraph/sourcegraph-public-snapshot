import assert from 'assert'

import { render } from '@testing-library/react'
import { createMemoryHistory } from 'history'
import * as sinon from 'sinon'
import { describe, expect, it } from 'vitest'

import { Link } from '../components/Link/Link'

import { createLinkClickHandler } from './linkClickHandler'

if (!global.SVGAElement) {
    // jsdom does not define SVGAElement, which is currently used in createLinkClickHandler. See
    // https://github.com/jsdom/jsdom/issues/2128.
    ;(global as any).SVGAElement = HTMLAnchorElement
}

describe('createLinkClickHandler', () => {
    it('handles clicks on links that stay inside the app', () => {
        // Create a URL that refers to the same host as `window.location.href`.
        const urlInsideApp = new URL('/else/where', window.location.href)
        expect(urlInsideApp.toString()).toBe(window.location.href + 'else/where')

        const history = createMemoryHistory({ initialEntries: [] })
        expect(history).toHaveLength(0)

        const { container } = render(
            // eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions
            <div onClick={createLinkClickHandler(history)}>
                <Link to={urlInsideApp.toString()}>Test</Link>
            </div>
        )

        const anchor = container.querySelector('a')
        assert(anchor)

        const spy = sinon.spy((_event: MouseEvent) => undefined)
        window.addEventListener('click', spy)

        anchor.click()

        sinon.assert.calledOnce(spy)
        expect(spy.args[0][0].defaultPrevented).toBe(true)

        expect(history).toHaveLength(1)
        expect(history.entries[0].pathname).toBe('/else/where')
    })

    it('ignores clicks on links that go outside the app', () => {
        // Create a URL that refers to a different host than `window.location.href`.
        const urlOutsideApp = new URL('https://other.example.com/some/where')
        expect(urlOutsideApp.origin).not.toBe(window.location.origin)

        const history = createMemoryHistory({ initialEntries: [] })
        expect(history).toHaveLength(0)

        const { container } = render(
            // eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions
            <div onClick={createLinkClickHandler(history)}>
                <Link to={urlOutsideApp.toString()}>Test</Link>
            </div>
        )

        const anchor = container.querySelector('a')
        assert(anchor)

        const spy = sinon.spy((_event: MouseEvent) => undefined)
        window.addEventListener('click', spy)

        anchor.click()

        sinon.assert.calledOnce(spy)
        expect(spy.args[0][0].defaultPrevented).toBe(false)
        expect(history).toHaveLength(0)
    })
})
