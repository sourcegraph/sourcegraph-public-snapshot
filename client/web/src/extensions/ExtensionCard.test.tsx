import { render } from '@testing-library/react'
import { MemoryRouter } from 'react-router'
import { CompatRouter } from 'react-router-dom-v5-compat'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import { AuthenticatedUser } from '../auth'

import { ExtensionCard } from './ExtensionCard'

const typescriptRawManifest = JSON.stringify({
    activationEvents: ['*'],
    browserslist: [],
    categories: ['Programming languages'],
    contributes: {},
    description: 'TypeScript/JavaScript code graph',
    devDependencies: {},
    extensionID: 'sourcegraph/typescript',
    main: 'dist/extension.js',
    name: 'typescript',
    publisher: 'sourcegraph',
    readme:
        '# Codecov Sourcegraph extension \n\n A [Sourcegraph extension](https://docs.sourcegraph.com/extensions) for showing code coverage information from [Codecov](https://codecov.io) on GitHub, Sourcegraph, and other tools. \n\n ## Features \n\n - Support for GitHub.com and Sourcegraph.com \n - Line coverage overlays on files (with green/yellow/red background colors) \n - Line branches/hits annotations on files \n - File coverage ratio indicator (`Coverage: N%`) and toggle button \n - Support for using a Codecov API token to see coverage for private repositories \n - File and directory coverage decorations on Sourcegraph \n\n ## Usage \n\n ### On GitHub using the Chrome extension \n 1. Install [Sourcegraph for Chrome](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack) \n 2. [Enable the Codecov extension on Sourcegraph](https://sourcegraph.com/extensions/sourcegraph/codecov) \n 3. Visit [tuf_store.go in theupdateframework/notary on GitHub](https://github.com/theupdateframework/notary/blob/master/server/storage/tuf_store.go) (or any other file in a public repository that has Codecov code coverage) \n 4. Click the `Coverage: N%` button to toggle Codecov test coverage background colors on the file (scroll down if they aren’t immediately visible) \n\n',
    scripts: {},
    tags: [],
    url:
        'https://sourcegraph.com/-/static/extension/7895-sourcegraph-typescript.js?c4ri18f15tv4--sourcegraph-typescript',
    version: '0.0.0-DEVELOPMENT',
})

const registryExtensionNode = {
    id: 'test-extension-1',
    publisher: null,
    extensionID: 'sourcegraph/typescript',
    extensionIDWithoutRegistry: 'sourcegraph/typescript',
    name: 'typescript',
    manifest: {
        raw: typescriptRawManifest,
        description: 'TypeScript/JavaScript code graph',
    },
    createdAt: '2019-01-26T03:39:17Z',
    updatedAt: '2019-01-26T03:39:17Z',
    url: '/extensions/sourcegraph/typescript',
    remoteURL: 'https://sourcegraph.com/extensions/sourcegraph/typescript',
    registryName: 'sourcegraph.com',
    isLocal: false,
    isWorkInProgress: false,
    viewerCanAdminister: false,
}

describe('ExtensionCard', () => {
    const NOOP_PLATFORM_CONTEXT: PlatformContext = {} as any

    const mockUser = {
        id: 'userID',
        username: 'username',
        email: 'user@me.com',
        siteAdmin: true,
    } as AuthenticatedUser

    test('renders', () => {
        expect(
            render(
                <MemoryRouter>
                    <CompatRouter>
                        <ExtensionCard
                            node={{
                                id: 'x/y',
                                manifest: {
                                    activationEvents: ['*'],
                                    description: 'd',
                                    url: 'https://example.com',
                                    icon: 'data:image/png,abcd',
                                },
                                registryExtension: registryExtensionNode,
                            }}
                            subject={{ id: 'u', viewerCanAdminister: false }}
                            viewerSubject={{
                                __typename: 'User',
                                username: 'u',
                                displayName: 'u',
                                id: 'u',
                                viewerCanAdminister: false,
                            }}
                            siteSubject={{
                                __typename: 'Site',
                                id: 's',
                                viewerCanAdminister: true,
                                allowSiteSettingsEdits: true,
                            }}
                            settingsCascade={{ final: null, subjects: null }}
                            platformContext={NOOP_PLATFORM_CONTEXT}
                            enabled={false}
                            enabledForAllUsers={false}
                            isLightTheme={false}
                            settingsURL="/settings/foobar"
                            authenticatedUser={mockUser}
                        />
                    </CompatRouter>
                </MemoryRouter>
            ).asFragment()
        ).toMatchSnapshot()
    })
})
