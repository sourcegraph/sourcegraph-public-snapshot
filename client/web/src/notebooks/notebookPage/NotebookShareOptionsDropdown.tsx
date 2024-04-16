import React, { type FC, useMemo } from 'react'

import { mdiChevronDown, mdiChevronUp, mdiDomain, mdiLock, mdiWeb } from '@mdi/js'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Icon, Menu, MenuButton, MenuItem, MenuList } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { OrgAvatar } from '../../org/OrgAvatar'

import styles from './NotebookShareOptionsDropdown.module.scss'

export interface ShareOption {
    namespaceType: 'User' | 'Org'
    namespaceName: string
    namespaceId: string
    isPublic: boolean
}

interface NotebookShareOptionsDropdownProps extends TelemetryProps, TelemetryV2Props {
    isSourcegraphDotCom: boolean
    authenticatedUser: AuthenticatedUser
    selectedShareOption: ShareOption
    onSelectShareOption: (shareOption: ShareOption) => void
}

const ShareOptionComponent: React.FunctionComponent<
    React.PropsWithChildren<Omit<ShareOption, 'namespaceId'> & { isSourcegraphDotCom: boolean }>
> = ({ isSourcegraphDotCom, namespaceType, namespaceName, isPublic }) => {
    if (namespaceType === 'User') {
        if (isPublic) {
            const publicText = isSourcegraphDotCom ? 'Public' : 'Instance'
            return (
                <>
                    <Icon
                        className="mr-2"
                        svgPath={isSourcegraphDotCom ? mdiWeb : mdiDomain}
                        inline={false}
                        height="1.15rem"
                        width="1.15rem"
                        aria-hidden={true}
                    />{' '}
                    {publicText}
                </>
            )
        }
        return (
            <>
                <Icon
                    className="mr-2"
                    svgPath={mdiLock}
                    inline={false}
                    aria-hidden={true}
                    height="1.15rem"
                    width="1.15rem"
                />{' '}
                Private
            </>
        )
    }
    return (
        <>
            <OrgAvatar org={namespaceName} className="d-inline-flex mr-2" size="sm" /> {namespaceName}
        </>
    )
}

export const NotebookShareOptionsDropdown: FC<NotebookShareOptionsDropdownProps> = props => {
    const {
        isSourcegraphDotCom,
        telemetryService,
        telemetryRecorder,
        authenticatedUser,
        selectedShareOption,
        onSelectShareOption,
    } = props

    const handleTriggerClick = (): void => {
        telemetryService.log('NotebookVisibilitySettingsDropdownToggled')
        telemetryRecorder.recordEvent('notebook.visibilitySettingsDropdown', 'toggle')
    }

    const shareOptions: ShareOption[] = useMemo(
        () => [
            {
                namespaceType: 'User' as const,
                isPublic: false,
                namespaceName: authenticatedUser.username,
                namespaceId: authenticatedUser.id,
            },
            ...authenticatedUser.organizations.nodes.map(org => ({
                namespaceType: 'Org' as const,
                isPublic: false,
                namespaceName: org.name,
                namespaceId: org.id,
            })),
            {
                namespaceType: 'User' as const,
                isPublic: true,
                namespaceName: authenticatedUser.username,
                namespaceId: authenticatedUser.id,
            },
        ],
        [authenticatedUser]
    )

    return (
        <Menu>
            <MenuButton
                outline={true}
                variant="secondary"
                data-testid="share-notebook-options-dropdown-toggle"
                className={styles.button}
                onClick={handleTriggerClick}
            >
                {isOpen => (
                    <>
                        <span className="d-flex align-items-center">
                            <ShareOptionComponent {...selectedShareOption} isSourcegraphDotCom={isSourcegraphDotCom} />
                        </span>
                        <span className="ml-5">
                            {isOpen ? (
                                <Icon svgPath={mdiChevronUp} inline={false} aria-hidden={true} />
                            ) : (
                                <Icon svgPath={mdiChevronDown} inline={false} aria-hidden={true} />
                            )}
                        </span>
                    </>
                )}
            </MenuButton>

            <MenuList
                // Stop keydown event bubbling in order to prevent a global notebook keyboard
                // navigation. Global notebook navigation breaks this menu keyboard navigation
                // see https://github.com/sourcegraph/sourcegraph/pull/41654#issuecomment-1246672813
                onKeyDown={event => event.stopPropagation()}
            >
                {shareOptions.map(option => (
                    <MenuItem
                        key={`${option.namespaceId}-${option.isPublic}`}
                        data-testid={`share-notebook-option-${option.namespaceName}-${option.isPublic}`}
                        className="d-flex align-items-center"
                        onSelect={() => onSelectShareOption(option)}
                    >
                        <ShareOptionComponent {...option} isSourcegraphDotCom={isSourcegraphDotCom} />
                    </MenuItem>
                ))}
            </MenuList>
        </Menu>
    )
}
