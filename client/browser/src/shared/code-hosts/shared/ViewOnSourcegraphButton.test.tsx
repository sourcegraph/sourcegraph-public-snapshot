import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'
import { noop } from 'lodash'

import { HTTPStatusError } from '@sourcegraph/http-client'

import { ViewOnSourcegraphButton } from './ViewOnSourcegraphButton'

describe('<ViewOnSourcegraphButton />', () => {
    describe('repository exists on the instance', () => {
        it('renders a link to the repository on the Sourcegraph instance', () => {
            expect(
                render(
                    <ViewOnSourcegraphButton
                        codeHostType="test-codehost"
                        sourcegraphURL="https://test.com"
                        userSettingsURL="https://sourcegraph.com/users/john.doe/settings/repositories/manage"
                        context={{ rawRepoName: 'test', privateRepository: false }}
                        className="test"
                        repoExistsOrError={true}
                        minimalUI={false}
                    />
                ).asFragment()
            ).toMatchSnapshot()
        })

        it('renders nothing in minimal UI mode', () => {
            expect(
                render(
                    <ViewOnSourcegraphButton
                        codeHostType="test-codehost"
                        sourcegraphURL="https://test.com"
                        userSettingsURL="https://sourcegraph.com/users/john.doe/settings/repositories/manage"
                        context={{ rawRepoName: 'test', privateRepository: false }}
                        className="test"
                        repoExistsOrError={true}
                        minimalUI={true}
                    />
                ).asFragment()
            ).toMatchSnapshot()
        })

        it('renders a link with the revision when provided', () => {
            expect(
                render(
                    <ViewOnSourcegraphButton
                        codeHostType="test-codehost"
                        sourcegraphURL="https://test.com"
                        userSettingsURL="https://sourcegraph.com/users/john.doe/settings/repositories/manage"
                        context={{
                            rawRepoName: 'test',
                            revision: 'test',
                            privateRepository: false,
                        }}
                        className="test"
                        repoExistsOrError={true}
                        minimalUI={false}
                    />
                ).asFragment()
            ).toMatchSnapshot()
        })
    })

    describe('repository does not exist on the instance', () => {
        it('renders "Configure Sourcegraph" button when pointing at sourcegraph.com', () => {
            expect(
                render(
                    <ViewOnSourcegraphButton
                        codeHostType="test-codehost"
                        sourcegraphURL="https://sourcegraph.com"
                        userSettingsURL="https://sourcegraph.com/users/john.doe/settings/repositories/manage"
                        context={{
                            rawRepoName: 'test',
                            revision: 'test',
                            privateRepository: false,
                        }}
                        className="test"
                        repoExistsOrError={false}
                        onConfigureSourcegraphClick={noop}
                        minimalUI={false}
                    />
                ).asFragment()
            ).toMatchSnapshot()
        })

        it('renders a "Repository not found" button when not pointing at sourcegraph.com', () => {
            expect(
                render(
                    <ViewOnSourcegraphButton
                        codeHostType="test-codehost"
                        sourcegraphURL="https://sourcegraph.test"
                        userSettingsURL="https://sourcegraph.com/users/john.doe/settings/repositories/manage"
                        context={{
                            rawRepoName: 'test',
                            revision: 'test',
                            privateRepository: false,
                        }}
                        className="test"
                        repoExistsOrError={false}
                        onConfigureSourcegraphClick={noop}
                        minimalUI={false}
                    />
                ).asFragment()
            ).toMatchSnapshot()
        })
    })

    describe('existence could not be determined ', () => {
        describe('because of an authentication failure', () => {
            for (const minimalUI of [true, false]) {
                describe(`minimalUI = ${String(minimalUI)}`, () => {
                    it('renders a sign in button if showSignInButton = true', () => {
                        expect(
                            render(
                                <ViewOnSourcegraphButton
                                    codeHostType="test-codehost"
                                    sourcegraphURL="https://test.com"
                                    userSettingsURL="https://sourcegraph.com/users/john.doe/settings/repositories/manage"
                                    context={{
                                        rawRepoName: 'test',
                                        revision: 'test',
                                        privateRepository: false,
                                    }}
                                    showSignInButton={true}
                                    className="test"
                                    repoExistsOrError={new HTTPStatusError(new Response('', { status: 401 }))}
                                    minimalUI={minimalUI}
                                />
                            ).asFragment()
                        ).toMatchSnapshot()
                    })
                })
            }
        })

        describe('because of an unknown error', () => {
            it('renders a button with an error label', () => {
                expect(
                    render(
                        <ViewOnSourcegraphButton
                            codeHostType="test-codehost"
                            sourcegraphURL="https://test.com"
                            userSettingsURL="https://sourcegraph.com/users/john.doe/settings/repositories/manage"
                            context={{
                                rawRepoName: 'test',
                                revision: 'test',
                                privateRepository: false,
                            }}
                            showSignInButton={true}
                            className="test"
                            repoExistsOrError={new Error('Something unknown happened!')}
                            minimalUI={false}
                        />
                    ).asFragment()
                ).toMatchSnapshot()
            })
        })
    })
})
