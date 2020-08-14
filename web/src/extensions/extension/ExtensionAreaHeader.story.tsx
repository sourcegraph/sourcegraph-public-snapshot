import * as H from 'history'
import { storiesOf } from '@storybook/react'
import React from 'react'
import webStyles from '../../SourcegraphWebApp.scss'
import { ExtensionAreaHeader, ExtensionAreaHeaderProps } from './ExtensionAreaHeader'
import * as GQL from '../../../../shared/src/graphql/schema'
import { ConfiguredRegistryExtension } from '../../../../shared/src/extensions/extension'
import { MemoryRouter } from 'react-router'

const { add } = storiesOf('web/ExtensionAreaHeader', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="layout__app-router-container">
            <div className="container mt-3">{story()}</div>
        </div>
    </>
))

const history = H.createMemoryHistory()

const icon =
    'data:image/png;base64,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'

const DEFAULT_EXTENSION: ConfiguredRegistryExtension<GQL.IRegistryExtension> = {
    id: 'sourcegraph/codecov',
    manifest: {
        url: 'a',
        name: 'Codecov',
        description: 'Shows code coverage from Codecov',
        activationEvents: ['*'],
        icon,
    },
    rawManifest: '{}',
    registryExtension: {
        id: 'sourcegraph/codecov',
        extensionIDWithoutRegistry: 'codecov',
        viewerCanAdminister: true,
    } as GQL.IRegistryExtension,
}

const DEFAULT_USER = {
    id: 'a',
    username: 'janedoe',
} as GQL.IUser

const DEFAULT_SETTINGS = {
    extensions: {
        'sourcegraph/codecov': true,
    },
}

const DEFAULT_PROPS: ExtensionAreaHeaderProps = {
    url: 'http://sourcegraph.com',
    extension: DEFAULT_EXTENSION,
    onDidUpdateExtension: () => {},
    authenticatedUser: DEFAULT_USER,
    navItems: [],
    isLightTheme: true,
    telemetryService: {
        log: () => {},
        logViewEvent: () => {},
    },
    history,
    location: history.location,
    platformContext: {
        updateSettings: () => Promise.resolve(),
    },
    match: {
        params: { id: '' },
        isExact: true,
        path: '',
        url: '',
    },
    settingsCascade: {
        subjects: [],
        final: DEFAULT_SETTINGS,
    },
}

add(
    'Default',
    () => (
        <MemoryRouter>
            <ExtensionAreaHeader {...DEFAULT_PROPS} />
        </MemoryRouter>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/wxYRk6fiYc7mSGLE1jvz1f/11774-Design-templates-for-common-page-types-Final-review?node-id=138%3A2213',
        },
    }
)

add(
    'Default - disabled',
    () => (
        <MemoryRouter>
            <ExtensionAreaHeader
                {...{
                    ...DEFAULT_PROPS,
                    settingsCascade: {
                        ...DEFAULT_PROPS.settingsCascade,
                        final: {
                            extensions: {},
                        },
                    },
                }}
            />
        </MemoryRouter>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/wxYRk6fiYc7mSGLE1jvz1f/11774-Design-templates-for-common-page-types-Final-review?node-id=192%3A1368',
        },
    }
)

add('Signed out - enabled', () => (
    <MemoryRouter>
        <ExtensionAreaHeader
            {...{
                ...DEFAULT_PROPS,
                authenticatedUser: null,
            }}
        />
    </MemoryRouter>
))

add('Signed out - disabled', () => (
    <MemoryRouter>
        <ExtensionAreaHeader
            {...{
                ...DEFAULT_PROPS,
                authenticatedUser: null,
                settingsCascade: {
                    ...DEFAULT_PROPS.settingsCascade,
                    final: {
                        extensions: {},
                    },
                },
            }}
        />
    </MemoryRouter>
))

add('No icon', () => (
    <MemoryRouter>
        <ExtensionAreaHeader
            {...{
                ...DEFAULT_PROPS,
                extension: {
                    ...DEFAULT_EXTENSION,
                    manifest: {
                        ...DEFAULT_EXTENSION.manifest,
                        icon: undefined,
                    },
                } as ConfiguredRegistryExtension<GQL.IRegistryExtension>,
            }}
        />
    </MemoryRouter>
))

add('Long description', () => (
    <MemoryRouter>
        <ExtensionAreaHeader
            {...{
                ...DEFAULT_PROPS,
                extension: {
                    ...DEFAULT_EXTENSION,
                    manifest: {
                        ...DEFAULT_EXTENSION.manifest,
                        description:
                            'A Sourcegraph extension that adds useful features when viewing files in a Git repository on Sourcegraph, GitHub, GitLab, and other supported code hosts.',
                    },
                } as ConfiguredRegistryExtension<GQL.IRegistryExtension>,
            }}
        />
    </MemoryRouter>
))

add('No description', () => (
    <MemoryRouter>
        <ExtensionAreaHeader
            {...{
                ...DEFAULT_PROPS,
                extension: {
                    ...DEFAULT_EXTENSION,
                    manifest: {
                        ...DEFAULT_EXTENSION.manifest,
                        description: undefined,
                    },
                } as ConfiguredRegistryExtension<GQL.IRegistryExtension>,
            }}
        />
    </MemoryRouter>
))

add('No icon, no description', () => (
    <MemoryRouter>
        <ExtensionAreaHeader
            {...{
                ...DEFAULT_PROPS,
                extension: {
                    ...DEFAULT_EXTENSION,
                    manifest: {
                        ...DEFAULT_EXTENSION.manifest,
                        description: undefined,
                        icon: undefined,
                    },
                } as ConfiguredRegistryExtension<GQL.IRegistryExtension>,
            }}
        />
    </MemoryRouter>
))

add('Dashed name', () => (
    <MemoryRouter>
        <ExtensionAreaHeader
            {...{
                ...DEFAULT_PROPS,
                extension: {
                    ...DEFAULT_EXTENSION,
                    manifest: {
                        ...DEFAULT_EXTENSION.manifest,
                        name: 'sourcegraph-npm-import-stats',
                    },
                } as ConfiguredRegistryExtension<GQL.IRegistryExtension>,
            }}
        />
    </MemoryRouter>
))

add('No name - falling back to id without registry', () => (
    <MemoryRouter>
        <ExtensionAreaHeader
            {...{
                ...DEFAULT_PROPS,
                extension: {
                    ...DEFAULT_EXTENSION,
                    manifest: {
                        ...DEFAULT_EXTENSION.manifest,
                        name: undefined,
                    },
                } as ConfiguredRegistryExtension<GQL.IRegistryExtension>,
            }}
        />
    </MemoryRouter>
))

add('No name, no registry extension - falling back to id', () => (
    <MemoryRouter>
        <ExtensionAreaHeader
            {...{
                ...DEFAULT_PROPS,
                extension: {
                    ...DEFAULT_EXTENSION,
                    manifest: {
                        ...DEFAULT_EXTENSION.manifest,
                        name: undefined,
                    },
                    registryExtension: undefined,
                } as ConfiguredRegistryExtension<GQL.IRegistryExtension>,
            }}
        />
    </MemoryRouter>
))
