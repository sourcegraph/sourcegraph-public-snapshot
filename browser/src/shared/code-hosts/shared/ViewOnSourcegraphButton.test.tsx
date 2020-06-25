import { ViewOnSourcegraphButton } from './ViewOnSourcegraphButton'
import { HTTPStatusError } from '../../../../../shared/src/backend/fetch'
import * as React from 'react'
import { noop } from 'lodash'
import { mount } from 'enzyme'

describe('<ViewOnSourcegraphButton />', () => {
    describe('repository exists on the instance', () => {
        it('renders a link to the repository on the Sourcegraph instance', () => {
            expect(
                mount(
                    <ViewOnSourcegraphButton
                        codeHostType="test-codehost"
                        sourcegraphURL="https://test.com"
                        getContext={() => ({ rawRepoName: 'test', privateRepository: false })}
                        className="test"
                        repoExistsOrError={true}
                        minimalUI={false}
                    />
                )
            ).toMatchSnapshot()
        })

        it('renders nothing in minimal UI mode', () => {
            expect(
                mount(
                    <ViewOnSourcegraphButton
                        codeHostType="test-codehost"
                        sourcegraphURL="https://test.com"
                        getContext={() => ({ rawRepoName: 'test', privateRepository: false })}
                        className="test"
                        repoExistsOrError={true}
                        minimalUI={true}
                    />
                )
            ).toMatchSnapshot()
        })

        it('renders a link with the revision when provided', () => {
            expect(
                mount(
                    <ViewOnSourcegraphButton
                        codeHostType="test-codehost"
                        sourcegraphURL="https://test.com"
                        getContext={() => ({
                            rawRepoName: 'test',
                            revision: 'test',
                            privateRepository: false,
                        })}
                        className="test"
                        repoExistsOrError={true}
                        minimalUI={false}
                    />
                )
            ).toMatchSnapshot()
        })
    })

    describe('repository does not exist on the instance', () => {
        it('renders "Configure Sourcegraph" button when pointing at sourcegraph.com', () => {
            expect(
                mount(
                    <ViewOnSourcegraphButton
                        codeHostType="test-codehost"
                        sourcegraphURL="https://sourcegraph.com"
                        getContext={() => ({
                            rawRepoName: 'test',
                            revision: 'test',
                            privateRepository: false,
                        })}
                        className="test"
                        repoExistsOrError={false}
                        onConfigureSourcegraphClick={noop}
                        minimalUI={false}
                    />
                )
            ).toMatchSnapshot()
        })

        it('renders a "Repository not found" button when not pointing at sourcegraph.com', () => {
            expect(
                mount(
                    <ViewOnSourcegraphButton
                        codeHostType="test-codehost"
                        sourcegraphURL="https://sourcegraph.test"
                        getContext={() => ({
                            rawRepoName: 'test',
                            revision: 'test',
                            privateRepository: false,
                        })}
                        className="test"
                        repoExistsOrError={false}
                        onConfigureSourcegraphClick={noop}
                        minimalUI={false}
                    />
                )
            ).toMatchSnapshot()
        })
    })

    describe('existence could not be determined ', () => {
        describe('because of an authentication failure', () => {
            for (const minimalUI of [true, false]) {
                describe(`minimalUI = ${String(minimalUI)}`, () => {
                    it('renders a sign in button if showSignInButton = true', () => {
                        expect(
                            mount(
                                <ViewOnSourcegraphButton
                                    codeHostType="test-codehost"
                                    sourcegraphURL="https://test.com"
                                    getContext={() => ({
                                        rawRepoName: 'test',
                                        revision: 'test',
                                        privateRepository: false,
                                    })}
                                    showSignInButton={true}
                                    className="test"
                                    repoExistsOrError={new HTTPStatusError(new Response('', { status: 401 }))}
                                    minimalUI={minimalUI}
                                />
                            )
                        ).toMatchSnapshot()
                    })
                })
            }
        })

        describe('because of an unknown error', () => {
            it('renders a button with an error label', () => {
                expect(
                    mount(
                        <ViewOnSourcegraphButton
                            codeHostType="test-codehost"
                            sourcegraphURL="https://test.com"
                            getContext={() => ({
                                rawRepoName: 'test',
                                revision: 'test',
                                privateRepository: false,
                            })}
                            showSignInButton={true}
                            className="test"
                            repoExistsOrError={new Error('Something unknown happened!')}
                            minimalUI={false}
                        />
                    )
                ).toMatchSnapshot()
            })
        })
    })
})
