import { of, throwError } from 'rxjs'
import { ViewOnSourcegraphButton } from './external_links'
import { HTTPStatusError } from '../../../../shared/src/backend/fetch'
import * as React from 'react'
import renderer, { ReactTestRenderer } from 'react-test-renderer'
import { noop } from 'lodash'

describe('<ViewOnSourcegraphButton />', () => {
    it('renders a link to the repository on the Sourcegraph instance', () => {
        let root: ReactTestRenderer
        renderer.act(() => {
            root = renderer.create(
                <ViewOnSourcegraphButton
                    sourcegraphURL="https://test.com"
                    context={{ rawRepoName: 'test', privateRepository: false }}
                    className="test"
                    ensureRepoExists={() => of(true)}
                    minimalUI={false}
                />
            )
        })
        expect(root!).toMatchSnapshot()
    })

    it('renders nothing in minimal UI mode', () => {
        let root: ReactTestRenderer
        renderer.act(() => {
            root = renderer.create(
                <ViewOnSourcegraphButton
                    sourcegraphURL="https://test.com"
                    context={{ rawRepoName: 'test', privateRepository: false }}
                    className="test"
                    ensureRepoExists={() => of(true)}
                    minimalUI={true}
                />
            )
        })
        expect(root!).toMatchSnapshot()
    })

    it('renders a link with the rev when provided', () => {
        let root: ReactTestRenderer
        renderer.act(() => {
            root = renderer.create(
                <ViewOnSourcegraphButton
                    sourcegraphURL="https://test.com"
                    context={{
                        rawRepoName: 'test',
                        rev: 'test',
                        privateRepository: false,
                    }}
                    className="test"
                    ensureRepoExists={() => of(true)}
                    minimalUI={false}
                />
            )
        })
        expect(root!).toMatchSnapshot()
    })

    for (const minimalUI of [true, false]) {
        describe(`minimalUI = ${String(minimalUI)}`, () => {
            it('renders a sign in button when authentication failed', () => {
                let root: ReactTestRenderer
                renderer.act(() => {
                    root = renderer.create(
                        <ViewOnSourcegraphButton
                            sourcegraphURL="https://test.com"
                            context={{
                                rawRepoName: 'test',
                                rev: 'test',
                                privateRepository: false,
                            }}
                            className="test"
                            ensureRepoExists={() => throwError(new HTTPStatusError(new Response('', { status: 401 })))}
                            minimalUI={minimalUI}
                        />
                    )
                })
                expect(root!).toMatchSnapshot()
            })
        })
    }

    it('renders configure sourcegraph button when pointing at sourcegraph.com and the repo does not exist', () => {
        let root: ReactTestRenderer
        renderer.act(() => {
            root = renderer.create(
                <ViewOnSourcegraphButton
                    sourcegraphURL="https://sourcegraph.com"
                    context={{
                        rawRepoName: 'test',
                        rev: 'test',
                        privateRepository: false,
                    }}
                    className="test"
                    ensureRepoExists={() => of(false)}
                    onConfigureSourcegraphClick={noop}
                    minimalUI={false}
                />
            )
        })
        expect(root!).toMatchSnapshot()
    })

    it('still renders a button to a private instance if repo does not exist', () => {
        let root: ReactTestRenderer
        renderer.act(() => {
            root = renderer.create(
                <ViewOnSourcegraphButton
                    sourcegraphURL="https://test.com"
                    context={{
                        rawRepoName: 'test',
                        rev: 'test',
                        privateRepository: false,
                    }}
                    className="test"
                    ensureRepoExists={() => of(false)}
                    minimalUI={false}
                />
            )
        })
        expect(root!).toMatchSnapshot()
    })
})
