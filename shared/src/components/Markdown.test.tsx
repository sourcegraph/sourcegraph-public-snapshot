import React from 'react'
import ReactDOM from 'react-dom'
import renderer from 'react-test-renderer'
import { Markdown } from './Markdown'
import { createMemoryHistory } from 'history'
import { renderMarkdown } from '../util/markdown'
import * as sinon from 'sinon'

describe('Markdown', () => {
    it('renders', () => {
        const history = createMemoryHistory()
        const component = renderer.create(<Markdown history={history} dangerousInnerHTML="hello" />)
        expect(component.toJSON()).toMatchSnapshot()
    })
    it('handles clicks on links that stay inside the app', () => {
        jsdom.reconfigure({ url: 'https://sourcegraph.test/some/where' })
        const root = document.body.appendChild(document.createElement('div'))
        const history = createMemoryHistory({ initialEntries: [] })
        expect(history).toHaveLength(0)
        ReactDOM.render(
            <Markdown
                history={history}
                dangerousInnerHTML={renderMarkdown('[test](https://sourcegraph.test/else/where)')}
            />,
            root
        )
        const anchor = root.querySelector('a')
        expect(anchor).toBeDefined()
        const spy = sinon.spy((_event: MouseEvent) => undefined)
        window.addEventListener('click', spy)
        anchor!.click()
        sinon.assert.calledOnce(spy)
        expect(spy.args[0][0].defaultPrevented).toBe(true)
        expect(history).toHaveLength(1)
        expect(history.entries[0].pathname).toBe('/else/where')
    })
    it('ignores clicks on links that go outside the app', () => {
        jsdom.reconfigure({ url: 'https://sourcegraph.test/some/where' })
        const root = document.body.appendChild(document.createElement('div'))
        const history = createMemoryHistory({ initialEntries: [] })
        expect(history).toHaveLength(0)
        ReactDOM.render(
            <Markdown history={history} dangerousInnerHTML={renderMarkdown('[test](https://github.com/some/where)')} />,
            root
        )
        const anchor = root.querySelector('a')
        expect(anchor).toBeDefined()
        const spy = sinon.spy((_event: MouseEvent) => undefined)
        window.addEventListener('click', spy)
        anchor!.click()
        sinon.assert.calledOnce(spy)
        expect(spy.args[0][0].defaultPrevented).toBe(false)
        expect(history).toHaveLength(0)
    })
})
