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
                    getContext={() => ({ rawRepoName: 'test', privateRepository: false })}
                    className="test"
                    repoExistsOrError={true}
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
                    getContext={() => ({ rawRepoName: 'test', privateRepository: false })}
                    className="test"
                    repoExistsOrError={true}
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
                    getContext={() => ({
                        rawRepoName: 'test',
                        rev: 'test',
                        privateRepository: false,
                    })}
                    className="test"
                    repoExistsOrError={true}
                    minimalUI={false}
                />
            )
        })
        expect(root!).toMatchSnapshot()
    })

    for (const minimalUI of [true, false]) {
        describe(`minimalUI = ${String(minimalUI)}`, () => {
            it('renders a sign in button when authentication failed and showSignInButton = true', () => {
                let root: ReactTestRenderer
                renderer.act(() => {
                    root = renderer.create(
                        <ViewOnSourcegraphButton
                            sourcegraphURL="https://test.com"
                            getContext={() => ({
                                rawRepoName: 'test',
                                rev: 'test',
                                privateRepository: false,
                            })}
                            showSignInButton={true}
                            className="test"
                            repoExistsOrError={new HTTPStatusError(new Response('', { status: 401 }))}
                            minimalUI={minimalUI}
                        />
                    )
                })
                expect(root!).toMatchSnapshot()
            })
        })
    }

    it('renders a button with an error label if the repo exists check failed with an unknown error', () => {
        let root: ReactTestRenderer
        renderer.act(() => {
            root = renderer.create(
                <ViewOnSourcegraphButton
                    sourcegraphURL="https://test.com"
                    getContext={() => ({
                        rawRepoName: 'test',
                        rev: 'test',
                        privateRepository: false,
                    })}
                    showSignInButton={true}
                    className="test"
                    repoExistsOrError={new Error('Something unknown happened!')}
                    minimalUI={false}
                />
            )
        })
        expect(root!).toMatchSnapshot()
    })

    it('renders configure sourcegraph button when pointing at sourcegraph.com and the repo does not exist', () => {
        let root: ReactTestRenderer
        renderer.act(() => {
            root = renderer.create(
                <ViewOnSourcegraphButton
                    sourcegraphURL="https://sourcegraph.com"
                    getContext={() => ({
                        rawRepoName: 'test',
                        rev: 'test',
                        privateRepository: false,
                    })}
                    className="test"
                    repoExistsOrError={false}
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
                    getContext={() => ({
                        rawRepoName: 'test',
                        rev: 'test',
                        privateRepository: false,
                    })}
                    className="test"
                    repoExistsOrError={false}
                    minimalUI={false}
                />
            )
        })
        expect(root!).toMatchSnapshot()
    })
})
