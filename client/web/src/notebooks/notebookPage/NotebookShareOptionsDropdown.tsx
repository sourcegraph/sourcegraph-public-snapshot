import React, { useCallback, useMemo } from 'react'

import { mdiLock, mdiChevronUp, mdiChevronDown, mdiDomain, mdiWeb } from '@mdi/js'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Menu, MenuButton, MenuItem, MenuList, Icon } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { OrgAvatar } from '../../org/OrgAvatar'

import styles from './NotebookShareOptionsDropdown.module.scss'

export interface ShareOption {
    namespaceType: 'User' | 'Org'
    namespaceName: string
    namespaceId: string
    isPublic: boolean
}

interface NotebookShareOptionsDropdownProps extends TelemetryProps {
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

export const NotebookShareOptionsDropdown: React.FunctionComponent<
    React.PropsWithChildren<NotebookShareOptionsDropdownProps>
> = ({ isSourcegraphDotCom, telemetryService, authenticatedUser, selectedShareOption, onSelectShareOption }) => {
    const handleToggle = useCallback(() => {
        telemetryService.log('NotebookVisibilitySettingsDropdownToggled')
    }, [telemetryService])

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
            {({ isOpen }) => (
                <>
                    <MenuButton
                        onClick={handleToggle}
                        className={styles.button}
                        outline={true}
                        variant="secondary"
                        data-testid="share-notebook-options-dropdown-toggle"
                    >
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
                    </MenuButton>
                    <MenuList>
                        {shareOptions.map(option => (
                            <MenuItem
                                key={`${option.namespaceId}-${option.isPublic}`}
                                className="d-flex align-items-center"
                                onSelect={() => onSelectShareOption(option)}
                                data-testid={`share-notebook-option-${option.namespaceName}-${option.isPublic}`}
                            >
                                <ShareOptionComponent {...option} isSourcegraphDotCom={isSourcegraphDotCom} />
                            </MenuItem>
                        ))}
                    </MenuList>
                </>
            )}
        </Menu>
    )
}
