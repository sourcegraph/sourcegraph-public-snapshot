import { fireEvent } from 'react-testing-library'
import { of } from 'rxjs'
import * as sinon from 'sinon'
import { injectViewContextOnSourcegraph } from './external_links'

describe('<ViewOnSourcegraphButton />', () => {
    beforeEach(() => {
        for (const test of document.querySelectorAll('.test')) {
            test.remove()
        }
    })

    it('renders a link', () => {
        injectViewContextOnSourcegraph(
            'https://test.com',
            {
                getContext: () => ({
                    repoName: 'test',
                }),
                getViewContextOnSourcegraphMount: () => {
                    const div = document.createElement('div')
                    document.body.appendChild(div)
                    return div
                },
                contextButtonClassName: 'test',
            },
            () => of(true),
            undefined
        )

        const link = document.querySelector<HTMLAnchorElement>('.test')
        expect(link).toBeInstanceOf(HTMLAnchorElement)
        expect(link!.href).toBe('https://test.com/test')
    })

    it('renders a link with the rev when provided', () => {
        injectViewContextOnSourcegraph(
            'https://test.com',
            {
                getContext: () => ({
                    repoName: 'test',
                    rev: 'test',
                }),
                getViewContextOnSourcegraphMount: () => {
                    const div = document.createElement('div')
                    document.body.appendChild(div)
                    return div
                },
                contextButtonClassName: 'test',
            },
            () => of(true),
            undefined
        )

        const link = document.querySelector<HTMLAnchorElement>('.test')
        expect(link).toBeInstanceOf(HTMLAnchorElement)
        expect(link!.href).toBe('https://test.com/test@test')
    })

    it('renders configure sourcegraph button when pointing at sourcegraph.com', () => {
        const configureClickSpy = sinon.spy()

        injectViewContextOnSourcegraph(
            'https://sourcegraph.com',
            {
                getContext: () => ({
                    repoName: 'test',
                    rev: 'test',
                }),
                getViewContextOnSourcegraphMount: () => {
                    const div = document.createElement('div')
                    document.body.appendChild(div)
                    return div
                },
                contextButtonClassName: 'test',
            },
            () => of(false),
            configureClickSpy
        )

        const link = document.querySelector<HTMLAnchorElement>('.test')
        expect(link).toBeInstanceOf(HTMLAnchorElement)
        expect(link!.textContent).toBe('Configure Sourcegraph')

        fireEvent.click(link!)
        expect(configureClickSpy.calledOnce).toBe(true)
    })

    it('still renders "View Repository" if repo doesn\'t exist and its not pointed at .com', () => {
        const configureClickSpy = sinon.spy()

        injectViewContextOnSourcegraph(
            'https://test.com',
            {
                getContext: () => ({
                    repoName: 'test',
                    rev: 'test',
                }),
                getViewContextOnSourcegraphMount: () => {
                    const div = document.createElement('div')
                    document.body.appendChild(div)
                    return div
                },
                contextButtonClassName: 'test',
            },
            () => of(false),
            configureClickSpy
        )

        const link = document.querySelector<HTMLAnchorElement>('.test')
        expect(link).toBeInstanceOf(HTMLAnchorElement)
        expect(link!.textContent).toBe('View Repository')
    })
})
