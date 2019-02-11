import * as H from 'history'
import { flatten } from 'lodash'
import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { setLinkComponent } from '../../../shared/src/components/Link'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { KeybindingsProps } from '../keybindings'
import { NavLinks } from './NavLinks'

// Renders a human-readable list of the NavLinks' contents so that humans can more easily diff
// snapshots to see what actually changed.
const renderShallow = (element: React.ReactElement<NavLinks['props']>): any => {
    const renderer = createRenderer()
    renderer.render(element)

    const getDisplayName = (element: React.ReactChild): string | string[] => {
        if (element === null) {
            return []
        } else if (typeof element === 'string' || typeof element === 'number') {
            return element.toString()
        } else if (element.type === 'li' && (element.props.children.props.href || element.props.children.props.to)) {
            return `${element.props.children.props.children} ${element.props.children.props.href ||
                element.props.children.props.to}`
        } else if (typeof element.type === 'symbol' || typeof element.type === 'string') {
            return flatten(React.Children.map(element.props.children, element => getDisplayName(element)))
        } else {
            return element.type.displayName || element.type.name || 'Unknown'
        }
    }

    return flatten(
        React.Children.map(renderer.getRenderOutput().props.children, e => getDisplayName(e)).filter(e => !!e)
    )
}

describe('NavLinks', () => {
    setLinkComponent((props: any) => <a {...props} />)
    afterAll(() => setLinkComponent(null as any)) // reset global env for other tests
    const NOOP_EXTENSIONS_CONTROLLER: ExtensionsControllerProps<
        'executeCommand' | 'services'
    >['extensionsController'] = { executeCommand: async () => void 0, services: {} as any }
    const NOOP_PLATFORM_CONTEXT = { forceUpdateTooltip: () => void 0 }
    const KEYBINDINGS: KeybindingsProps['keybindings'] = { commandPalette: [] }
    const SETTINGS_CASCADE: SettingsCascadeProps['settingsCascade'] = { final: null, subjects: null }
    // tslint:disable-next-line:no-object-literal-type-assertion
    const USER = { username: 'u' } as GQL.IUser
    const history = H.createMemoryHistory({ keyLength: 0 })
    const commonProps = {
        extensionsController: NOOP_EXTENSIONS_CONTROLLER,
        platformContext: NOOP_PLATFORM_CONTEXT,
        isLightTheme: true,
        onThemeChange: () => void 0,
        keybindings: KEYBINDINGS,
        settingsCascade: SETTINGS_CASCADE,
    }

    // The 3 main props that affect the desired contents of NavLinks are whether the user is signed
    // in, whether we're on Sourcegraph.com, and the path. Create snapshots of all permutations.
    for (const authenticatedUser of [null, USER]) {
        for (const showDotComMarketing of [false, true]) {
            for (const path of ['/foo', '/search', '/welcome']) {
                const name = [
                    authenticatedUser ? 'authed' : 'unauthed',
                    showDotComMarketing ? 'Sourcegraph.com' : 'self-hosted',
                    path,
                ].join(' ')
                test(name, () => {
                    expect(
                        renderShallow(
                            <NavLinks
                                {...commonProps}
                                authenticatedUser={authenticatedUser}
                                showDotComMarketing={showDotComMarketing}
                                location={H.createLocation(path, history.location)}
                            />
                        )
                    ).toMatchSnapshot()
                })
            }
        }
    }
})
