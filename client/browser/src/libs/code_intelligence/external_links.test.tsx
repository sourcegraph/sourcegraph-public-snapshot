import { fireEvent } from 'react-testing-library'
import { of } from 'rxjs'
import * as sinon from 'sinon'
import { renderViewContextOnSourcegraph } from './external_links'

describe('<ViewOnSourcegraphButton />', () => {
    let mount: HTMLElement
    beforeEach(() => {
        document.body.innerHTML = ''
        mount = document.createElement('div')
        document.body.append(mount)
    })

    it('renders a link', () => {
        renderViewContextOnSourcegraph({
            sourcegraphURL: 'https://test.com',
            getContext: () => ({ repoName: 'test', privateRepository: false }),
            viewOnSourcegraphButtonClassProps: {
                className: 'test',
            },
            ensureRepoExists: () => of(true),
        })(mount)

        const link = document.querySelector<HTMLAnchorElement>('.test')
        expect(link).toBeInstanceOf(HTMLAnchorElement)
        expect(link!.href).toBe('https://test.com/test')
    })

    it('renders a link with the rev when provided', () => {
        renderViewContextOnSourcegraph({
            sourcegraphURL: 'https://test.com',
            getContext: () => ({
                repoName: 'test',
                rev: 'test',
                privateRepository: false,
            }),
            viewOnSourcegraphButtonClassProps: {
                className: 'test',
            },
            ensureRepoExists: () => of(true),
        })(mount)

        const link = document.querySelector<HTMLAnchorElement>('.test')
        expect(link).toBeInstanceOf(HTMLAnchorElement)
        expect(link!.href).toBe('https://test.com/test@test')
    })

    it('renders configure sourcegraph button when pointing at sourcegraph.com', () => {
        const configureClickSpy = sinon.spy()

        renderViewContextOnSourcegraph({
            sourcegraphURL: 'https://sourcegraph.com',
            getContext: () => ({
                repoName: 'test',
                rev: 'test',
                privateRepository: false,
            }),
            viewOnSourcegraphButtonClassProps: {
                className: 'test',
            },
            ensureRepoExists: () => of(false),
            onConfigureSourcegraphClick: configureClickSpy,
        })(mount)

        const link = document.querySelector<HTMLAnchorElement>('.test')
        expect(link).toBeInstanceOf(HTMLAnchorElement)
        expect(link!.textContent).toBe(' Configure Sourcegraph')

        fireEvent.click(link!)
        expect(configureClickSpy.calledOnce).toBe(true)
    })

    it("still renders if repo doesn't exist and its not pointed at .com", () => {
        const configureClickSpy = sinon.spy()

        renderViewContextOnSourcegraph({
            sourcegraphURL: 'https://test.com',
            getContext: () => ({
                repoName: 'test',
                rev: 'test',
                privateRepository: false,
            }),
            viewOnSourcegraphButtonClassProps: {
                className: 'test',
            },
            ensureRepoExists: () => of(false),
            onConfigureSourcegraphClick: configureClickSpy,
        })(mount)

        const link = document.querySelector<HTMLAnchorElement>('.test')
        expect(link).toBeInstanceOf(HTMLAnchorElement)
        expect(link!.getAttribute('aria-label')).toBe('View repository on Sourcegraph')
    })
})
