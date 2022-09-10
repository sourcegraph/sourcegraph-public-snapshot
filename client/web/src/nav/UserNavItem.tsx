import React, { useCallback, useMemo } from 'react'

import { mdiChevronDown, mdiChevronUp, mdiOpenInNew } from '@mdi/js'
import { Shortcut } from '@slimsag/react-shortcuts'
import classNames from 'classnames'

import { useKeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts/useKeyboardShortcut'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    Menu,
    MenuButton,
    MenuDivider,
    MenuHeader,
    MenuItem,
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
import { ThemePreferenceProps, ThemePreference } from '../theme'
import { UserAvatar } from '../user/UserAvatar'

import styles from './UserNavItem.module.scss'

export interface UserNavItemProps extends ThemeProps, ThemePreferenceProps {
    authenticatedUser: Pick<
        AuthenticatedUser,
        'username' | 'avatarURL' | 'settingsURL' | 'organizations' | 'siteAdmin' | 'session' | 'displayName'
    >
    showDotComMarketing: boolean
    codeHostIntegrationMessaging: 'browser-extension' | 'native-integration'
    position?: Position
    menuButtonRef?: React.Ref<HTMLButtonElement>
    showKeyboardShortcutsHelp: () => void
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
    const keyboardShortcutSwitchTheme = useKeyboardShortcut('switchTheme')

    return (
        <>
            {keyboardShortcutSwitchTheme?.keybindings.map((keybinding, index) => (
                // `Shortcut` doesn't update its states when `onMatch` changes
                // so we put `themePreference` in `key` binding to make it
                <Shortcut key={`${themePreference}-${index}`} {...keybinding} onMatch={onThemeCycle} />
            ))}
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
                                    <Icon svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} aria-hidden={true} />
                                </div>
                            </div>
                        </MenuButton>

                        <MenuList position={position} className={styles.dropdownMenu} aria-label="User. Open menu">
                            <MenuHeader className={styles.dropdownHeader}>
                                Signed in as <strong>@{props.authenticatedUser.username}</strong>
                            </MenuHeader>
                            <MenuDivider className={styles.dropdownDivider} />
                            <MenuLink as={Link} to={props.authenticatedUser.settingsURL!}>
                                Settings
                            </MenuLink>
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
                            </div>
                            {!isOpenBetaEnabled && props.authenticatedUser.organizations.nodes.length > 0 && (
                                <>
                                    <MenuDivider className={styles.dropdownDivider} />
                                    <MenuHeader className={styles.dropdownHeader}>Your organizations</MenuHeader>
                                    {props.authenticatedUser.organizations.nodes.map(org => (
                                        <MenuLink as={Link} key={org.id} to={org.settingsURL || org.url}>
                                            {org.displayName || org.name}
                                        </MenuLink>
                                    ))}
                                </>
                            )}
                            <MenuDivider className={styles.dropdownDivider} />
                            {props.authenticatedUser.siteAdmin && (
                                <MenuLink as={Link} to="/site-admin">
                                    Site admin
                                </MenuLink>
                            )}
                            <MenuLink as={Link} to="/help" target="_blank" rel="noopener">
                                Help <Icon aria-hidden={true} svgPath={mdiOpenInNew} />
                            </MenuLink>
                            <MenuItem onSelect={props.showKeyboardShortcutsHelp}>Keyboard shortcuts</MenuItem>

                            {props.authenticatedUser.session?.canSignOut && (
                                <MenuLink as={AnchorLink} to="/-/sign-out">
                                    Sign out
                                </MenuLink>
                            )}
                            <MenuDivider className={styles.dropdownDivider} />
                            {props.showDotComMarketing && (
                                <MenuLink
                                    as={AnchorLink}
                                    to="https://about.sourcegraph.com"
                                    target="_blank"
                                    rel="noopener"
                                >
                                    About Sourcegraph <Icon aria-hidden={true} svgPath={mdiOpenInNew} />
                                </MenuLink>
                            )}
                            {codeHostIntegrationMessaging === 'browser-extension' && (
                                <MenuLink
                                    as={AnchorLink}
                                    to="/help/integration/browser_extension"
                                    target="_blank"
                                    rel="noopener"
                                >
                                    Browser extension <Icon aria-hidden={true} svgPath={mdiOpenInNew} />
                                </MenuLink>
                            )}
                        </MenuList>
                    </>
                )}
            </Menu>
        </>
    )
}
