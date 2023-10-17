import { type FC, useState } from 'react'

import { mdiChevronDown, mdiChevronUp, mdiRefresh } from '@mdi/js'
import classNames from 'classnames'

import {
    Icon,
    Link,
    Menu,
    MenuButton,
    MenuDivider,
    MenuHeader,
    MenuLink,
    MenuList,
    Position,
} from '@sourcegraph/wildcard'

import { InstallModal } from './InstallModal'
import { ChangelogModal } from './ReviewAndInstallModal'
import { type UpdateInfo, useUpdater } from './updater'

import styles from './UpdateGlobalNav.module.scss'

export interface UpdateGlobalNavFrameProps {
    details: UpdateInfo
}

export const UpdateGlobalNavFrame: FC<UpdateGlobalNavFrameProps> = ({ details }) => {
    const [showChangelog, setShowChangelog] = useState<boolean>(false)
    const [install, setInstall] = useState<boolean>(false)

    const onClose = (): void => {
        setInstall(false)
        setShowChangelog(false)
        details.checkNow?.(true)
    }

    return details.hasNewVersion ? (
        <>
            <Menu>
                {({ isExpanded }) => (
                    <>
                        <MenuButton
                            variant="link"
                            data-testid="update-nav-item-toggle"
                            className={classNames('d-flex align-items-center text-decoration-none', styles.menuButton)}
                            aria-label={`${isExpanded ? 'Close' : 'Open'} update menu`}
                        >
                            <div className="position-relative">
                                <div className="align-items-center d-flex">
                                    <Icon svgPath={mdiRefresh} aria-hidden={true} />
                                    Update Available
                                    <Icon svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} aria-hidden={true} />
                                </div>
                            </div>
                        </MenuButton>

                        <MenuList
                            position={Position.bottomEnd}
                            className={styles.dropdownMenu}
                            aria-label="User. Open menu"
                        >
                            <MenuHeader className={styles.dropdownHeader}>
                                <strong>{details.newVersion}</strong> Version is available
                            </MenuHeader>
                            <MenuDivider className={styles.dropdownDivider} />
                            <MenuLink
                                as={Link}
                                to=""
                                onClick={() => {
                                    setInstall(true)
                                    details.startInstall?.()
                                }}
                            >
                                Review and Install
                            </MenuLink>
                            <MenuLink as={Link} to="" onClick={() => setShowChangelog(true)}>
                                Changelog
                            </MenuLink>
                        </MenuList>
                    </>
                )}
            </Menu>
            {install && <InstallModal details={details} onClose={onClose} />}
            {showChangelog && <ChangelogModal details={details} fromSettingsPage={false} onClose={onClose} />}
        </>
    ) : (
        <></>
    )
}

export function UpdateGlobalNav(): JSX.Element {
    const update = useUpdater()
    return <UpdateGlobalNavFrame details={update} />
}
