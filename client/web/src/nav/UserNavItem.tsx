import React, { useCallback, useMemo, useState } from 'react'

import { Shortcut } from '@slimsag/react-shortcuts'
import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
// eslint-disable-next-line no-restricted-imports
import { Tooltip } from 'reactstrap'

import { KeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts'
import { KEYBOARD_SHORTCUT_SHOW_HELP } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    Menu,
    MenuButton,
    MenuDivider,
    MenuHeader,
    MenuItem,
    useTimeoutManager,
    MenuLink,
    MenuList,
    Link,
    Position,
    AnchorLink,
    Select,
    Icon,
    Badge,
} from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import { ThemePreference } from '../stores/themeState'
import { ThemePreferenceProps } from '../theme'
import { UserAvatar } from '../user/UserAvatar'

import styles from './UserNavItem.module.scss'

export interface UserNavItemProps extends ThemeProps, ThemePreferenceProps, ExtensionAlertAnimationProps {
    authenticatedUser: Pick<
        AuthenticatedUser,
        'username' | 'avatarURL' | 'settingsURL' | 'organizations' | 'siteAdmin' | 'session' | 'displayName'
    >
    showDotComMarketing: boolean
    keyboardShortcutForSwitchTheme?: KeyboardShortcut
    codeHostIntegrationMessaging: 'browser-extension' | 'native-integration'
    showRepositorySection?: boolean
    position?: Position
    menuButtonRef?: React.Ref<HTMLButtonElement>
}

export interface ExtensionAlertAnimationProps {
    isExtensionAlertAnimating: boolean
}

/**
 * React hook to manage the animation that occurs after the user dismisses
 * `InstallBrowserExtensionAlert`.
 *
 * This hook is called from the the LCA of `UserNavItem` and the component that triggers
 * the animation.
 */
export function useExtensionAlertAnimation(): ExtensionAlertAnimationProps & {
    startExtensionAlertAnimation: () => void
} {
    const [isAnimating, setIsAnimating] = useState(false)

    const animationManager = useTimeoutManager()

    const startExtensionAlertAnimation = useCallback(() => {
        if (!isAnimating) {
            setIsAnimating(true)

            animationManager.setTimeout(() => {
                setIsAnimating(false)
            }, 5100)
        }
    }, [isAnimating, animationManager])

    return { isExtensionAlertAnimating: isAnimating, startExtensionAlertAnimation }
}

/**
 * Triggers Keyboard Shortcut help when the button is clicked in the Menu Nav item
 */

const showKeyboardShortcutsHelp = (): void => {
    const keybinding = KEYBOARD_SHORTCUT_SHOW_HELP.keybindings[0]
    const shiftKey = !!keybinding.held?.includes('Shift')
    const altKey = !!keybinding.held?.includes('Alt')
    const metaKey = !!keybinding.held?.includes('Meta')
    const ctrlKey = !!keybinding.held?.includes('Control')

    for (const key of keybinding.ordered) {
        document.dispatchEvent(new KeyboardEvent('keydown', { key, shiftKey, metaKey, ctrlKey, altKey }))
    }
}

/**
 * Displays the user's avatar and/or username in the navbar and exposes a dropdown menu with more options for
 * authenticated viewers.
 */
export const UserNavItem: React.FunctionComponent<React.PropsWithChildren<UserNavItemProps>> = props => {
    const {
        menuButtonRef,
        themePreference,
        onThemePreferenceChange,
        isExtensionAlertAnimating,
        codeHostIntegrationMessaging,
        position = Position.bottomEnd,
    } = props

    const supportsSystemTheme = useMemo(
        () => Boolean(window.matchMedia?.('not all and (prefers-color-scheme), (prefers-color-scheme)').matches),
        []
    )

    const onThemeChange: React.ChangeEventHandler<HTMLSelectElement> = useCallback(
        event => {
            onThemePreferenceChange(event.target.value as ThemePreference)
        },
        [onThemePreferenceChange]
    )

    const onThemeCycle = useCallback((): void => {
        onThemePreferenceChange(themePreference === ThemePreference.Dark ? ThemePreference.Light : ThemePreference.Dark)
    }, [onThemePreferenceChange, themePreference])

    // Target ID for tooltip
    const targetID = 'target-user-avatar'
    const [isOpenBetaEnabled] = useFeatureFlag('open-beta-enabled')

    return (
        <Menu>
            {({ isExpanded }) => (
                <>
                    <MenuButton
                        ref={menuButtonRef}
                        variant="link"
                        data-testid="user-nav-item-toggle"
                        className={classNames('d-flex align-items-center text-decoration-none', styles.menuButton)}
                        aria-label={`${isExpanded ? 'Close' : 'Open'} user profile menu`}
                    >
                        <div className="position-relative">
                            <div className="align-items-center d-flex">
                                <UserAvatar
                                    user={props.authenticatedUser}
                                    targetID={targetID}
                                    className={styles.avatar}
                                />
                                <Icon role="img" as={isExpanded ? ChevronUpIcon : ChevronDownIcon} aria-hidden={true} />
                            </div>
                        </div>
                        {isExtensionAlertAnimating && (
                            <Tooltip
                                target={targetID}
                                placement="bottom"
                                isOpen={true}
                                modifiers={{
                                    offset: {
                                        offset: '0, 10px',
                                    },
                                }}
                                className={styles.tooltip}
                            >
                                Install the browser extension from here later
                            </Tooltip>
                        )}
                    </MenuButton>
                    <MenuList position={position} className={styles.dropdownMenu} aria-label="User. Open menu">
                        <MenuHeader>
                            Signed in as <strong>@{props.authenticatedUser.username}</strong>
                        </MenuHeader>
                        <MenuDivider />
                        <MenuLink as={Link} to={props.authenticatedUser.settingsURL!}>
                            Settings
                        </MenuLink>
                        {props.showRepositorySection && (
                            <MenuLink as={Link} to={`/users/${props.authenticatedUser.username}/settings/repositories`}>
                                Your repositories
                            </MenuLink>
                        )}
                        <MenuLink as={Link} to={`/users/${props.authenticatedUser.username}/searches`}>
                            Saved searches
                        </MenuLink>
                        {isOpenBetaEnabled && (
                            <MenuLink
                                as={Link}
                                to={`/users/${props.authenticatedUser.username}/settings/organizations`}
                            >
                                Your organizations <Badge variant="info">NEW</Badge>
                            </MenuLink>
                        )}
                        <MenuDivider />
                        <div className="px-2 py-1">
                            <div className="d-flex align-items-center">
                                <div className="mr-2">Theme</div>
                                <Select
                                    aria-label=""
                                    isCustomStyle={true}
                                    selectSize="sm"
                                    data-testid="theme-toggle"
                                    onChange={onThemeChange}
                                    value={props.themePreference}
                                    className="mb-0 flex-1"
                                >
                                    <option value={ThemePreference.Light}>Light</option>
                                    <option value={ThemePreference.Dark}>Dark</option>
                                    <option value={ThemePreference.System}>System</option>
                                </Select>
                            </div>
                            {props.themePreference === ThemePreference.System && !supportsSystemTheme && (
                                <div className="text-wrap">
                                    <small>
                                        <AnchorLink
                                            to="https://caniuse.com/#feat=prefers-color-scheme"
                                            className="text-warning"
                                            target="_blank"
                                            rel="noopener noreferrer"
                                        >
                                            Your browser does not support the system theme.
                                        </AnchorLink>
                                    </small>
                                </div>
                            )}
                            {props.keyboardShortcutForSwitchTheme?.keybindings.map((keybinding, index) => (
                                <Shortcut key={index} {...keybinding} onMatch={onThemeCycle} />
                            ))}
                        </div>
                        {!isOpenBetaEnabled && props.authenticatedUser.organizations.nodes.length > 0 && (
                            <>
                                <MenuDivider />
                                <MenuHeader>Your organizations</MenuHeader>
                                {props.authenticatedUser.organizations.nodes.map(org => (
                                    <MenuLink as={Link} key={org.id} to={org.settingsURL || org.url}>
                                        {org.displayName || org.name}
                                    </MenuLink>
                                ))}
                            </>
                        )}
                        <MenuDivider />
                        {props.authenticatedUser.siteAdmin && (
                            <MenuLink as={Link} to="/site-admin">
                                Site admin
                            </MenuLink>
                        )}
                        <MenuLink as={Link} to="/help" target="_blank" rel="noopener">
                            Help <Icon role="img" as={OpenInNewIcon} aria-hidden={true} />
                        </MenuLink>
                        <MenuItem onSelect={showKeyboardShortcutsHelp}>Keyboard shortcuts</MenuItem>

                        {props.authenticatedUser.session?.canSignOut && (
                            <MenuLink as={AnchorLink} to="/-/sign-out">
                                Sign out
                            </MenuLink>
                        )}
                        <MenuDivider />
                        {props.showDotComMarketing && (
                            <MenuLink as={AnchorLink} to="https://about.sourcegraph.com" target="_blank" rel="noopener">
                                About Sourcegraph <Icon role="img" as={OpenInNewIcon} aria-hidden={true} />
                            </MenuLink>
                        )}
                        {codeHostIntegrationMessaging === 'browser-extension' && (
                            <MenuLink
                                as={AnchorLink}
                                to="https://docs.sourcegraph.com/integration/browser_extension"
                                target="_blank"
                                rel="noopener"
                            >
                                Browser extension <Icon role="img" as={OpenInNewIcon} aria-hidden={true} />
                            </MenuLink>
                        )}
                    </MenuList>
                </>
            )}
        </Menu>
    )
}
