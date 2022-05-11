import assert from 'assert'

import { createMemoryHistory } from 'history'
import ReactDOM from 'react-dom'
import * as sinon from 'sinon'

import { Link } from '@sourcegraph/wildcard'

import { createLinkClickHandler } from './linkClickHandler'

describe('createLinkClickHandler', () => {
    it('handles clicks on links that stay inside the app', () => {
        jsdom.reconfigure({ url: 'https://sourcegraph.test/some/where' })

        const history = createMemoryHistory({ initialEntries: [] })
        expect(history).toHaveLength(0)

        const root = document.createElement('div')
        document.body.append(root)
        ReactDOM.render(
            // eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions
            <div onClick={createLinkClickHandler(history)}>
                <Link to="https://sourcegraph.test/else/where">Test</Link>
            </div>,
            root
        )

        const anchor = root.querySelector('a')
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
        jsdom.reconfigure({ url: 'https://sourcegraph.test/some/where' })
        const history = createMemoryHistory({ initialEntries: [] })
        expect(history).toHaveLength(0)

        const root = document.createElement('div')
        document.body.append(root)
        ReactDOM.render(
            // eslint-disable-next-line jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions
            <div onClick={createLinkClickHandler(history)}>
                <Link to="https://github.com/some/where">Test</Link>
            </div>,
            root
        )

        const anchor = root.querySelector('a')
        assert(anchor)

        const spy = sinon.spy((_event: MouseEvent) => undefined)
        window.addEventListener('click', spy)

        anchor.click()

        sinon.assert.calledOnce(spy)
        expect(spy.args[0][0].defaultPrevented).toBe(false)
        expect(history).toHaveLength(0)
    })
})
