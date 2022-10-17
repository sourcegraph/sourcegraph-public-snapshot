import { createMemoryHistory } from 'history'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { RegistryExtensionOverviewPage } from './RegistryExtensionOverviewPage'

jest.mock('mdi-react/GithubIcon', () => 'GithubIcon')

describe('RegistryExtensionOverviewPage', () => {
    test('renders', () => {
        const history = createMemoryHistory()
        expect(
            renderWithBrandedContext(
                <RegistryExtensionOverviewPage
                    telemetryService={NOOP_TELEMETRY_SERVICE}
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
                    isLightTheme={true}
                />,
                { history }
            ).asFragment()
        ).toMatchSnapshot()
    })

    // Meant to prevent this regression: https://github.com/sourcegraph/sourcegraph/pull/15040
    test('renders with no tags', () => {
        const history = createMemoryHistory()
        expect(
            renderWithBrandedContext(
                <RegistryExtensionOverviewPage
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    extension={{
                        id: 'x',
                        rawManifest: '{}',
                        manifest: {
                            url: 'https://example.com',
                            activationEvents: ['*'],
                            categories: ['Programming languages', 'Other'],
                            tags: [],
                            readme: '**A**',
                            repository: {
                                url: 'https://github.com/foo/bar',
                                type: 'git',
                            },
                        },
                    }}
                    isLightTheme={true}
                />,
                { history }
            ).asFragment()
        ).toMatchSnapshot()
    })

    describe('categories', () => {
        test('filters out unrecognized categories', () => {
            const history = createMemoryHistory()
            const output = renderWithBrandedContext(
                <RegistryExtensionOverviewPage
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    extension={{
                        id: 'x',
                        rawManifest: '',
                        manifest: {
                            url: 'https://example.com',
                            activationEvents: ['*'],
                            categories: ['Programming languages', 'invalid', 'Other'],
                        },
                    }}
                    isLightTheme={true}
                />,
                { history }
            )
            expect(toText(output.getAllByTestId('test-registry-extension-categories'))).toEqual([
                'Programming languages',
                'Other' /* no 'invalid' */,
            ])
        })
    })
})

function toText(values: string[] | HTMLElement[] | NodeListOf<ChildNode>): string[] {
    const textNodes: string[] = []
    for (const value of values) {
        if (typeof value === 'string') {
            textNodes.push(value)
        } else if (value.hasChildNodes()) {
            textNodes.push(...toText(value.childNodes))
        } else if (value.textContent !== null) {
            textNodes.push(value.textContent)
        }
    }
    return textNodes
}
