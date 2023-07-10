import { FC } from 'react'

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

import { UpdateInfo, useUpdater } from './updater'

import styles from './UpdateGlobalNav.module.scss'

interface UpdateGlobalNavFrameProps {
    details: UpdateInfo
}

const showChangelog = () => {}

const UpdateGlobalNavFrame: FC<UpdateGlobalNavFrameProps> = ({ details }) =>
    details.hasNewVersion ? (
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
                        <MenuLink as={Link} to="" onClick={details.startInstall}>
                            Install and Restart
                        </MenuLink>
                        <MenuLink as={Link} to="" onClick={showChangelog}>
                            Changelog
                        </MenuLink>
                    </MenuList>
                </>
            )}
        </Menu>
    ) : (
        <></>
    )

export function UpdateGlobalNav(): JSX.Element {
    const update = useUpdater()
    return <UpdateGlobalNavFrame details={update} />
}
