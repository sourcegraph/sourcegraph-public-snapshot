import { noop } from 'lodash'
import React from 'react'
import { RegistryExtensionOverviewPage } from './RegistryExtensionOverviewPage'
import { PageTitle } from '../../components/PageTitle'
import { createMemoryHistory } from 'history'
import { Router } from 'react-router'
import { mount } from 'enzyme'

jest.mock('mdi-react/GithubIcon', () => 'GithubIcon')

describe('RegistryExtensionOverviewPage', () => {
    afterEach(() => {
        PageTitle.titleSet = false
    })
    test('renders', () => {
        const history = createMemoryHistory()
        expect(
            mount(
                <Router history={history}>
                    <RegistryExtensionOverviewPage
                        eventLogger={{ logViewEvent: noop }}
                        extension={{
                            id: 'x',
                            rawManifest: '{}',
                            manifest: {
                                url: 'https://example.com',
                                activationEvents: ['*'],
                                categories: ['Programming languages', 'Other'],
                                tags: ['T1', 'T2'],
                                readme: '**A**',
                                repository: {
                                    url: 'https://github.com/foo/bar',
                                    type: 'git',
                                },
                            },
                        }}
                        history={history}
                    />
                </Router>
            ).children()
        ).toMatchSnapshot()
    })

    describe('categories', () => {
        test('filters out unrecognized categories', () => {
            const history = createMemoryHistory()
            const output = mount(
                <Router history={history}>
                    <RegistryExtensionOverviewPage
                        eventLogger={{ logViewEvent: noop }}
                        extension={{
                            id: 'x',
                            rawManifest: '',
                            manifest: {
                                url: 'https://example.com',
                                activationEvents: ['*'],
                                categories: ['Programming languages', 'invalid', 'Other'],
                            },
                        }}
                        history={createMemoryHistory()}
                    />
                </Router>
            )
            const foundCategories: string[] = []
            output.find('.list-inline-item').forEach(element => {
                foundCategories.push(element.getDOMNode().textContent!)
            })
            expect(foundCategories).toEqual(['Other', 'Programming languages' /* no 'invalid' */])
        })
    })
})
