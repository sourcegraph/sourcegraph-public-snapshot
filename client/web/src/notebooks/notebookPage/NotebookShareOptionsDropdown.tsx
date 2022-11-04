import React, { useCallback, useMemo, useState } from 'react'

import { mdiLock, mdiWeb, mdiDomain, mdiChevronUp, mdiChevronDown } from '@mdi/js'
// eslint-disable-next-line no-restricted-imports
import { Dropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Icon } from '@sourcegraph/wildcard'

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
    const [isOpen, setIsOpen] = useState(false)
    const handleToggle = useCallback(() => {
        telemetryService.log('NotebookVisibilitySettingsDropdownToggled')
        setIsOpen(isOpen => !isOpen)
    }, [telemetryService, setIsOpen])

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
        <Dropdown isOpen={isOpen} toggle={handleToggle}>
            <DropdownToggle
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
            </DropdownToggle>
            <DropdownMenu>
                {shareOptions.map(option => (
                    <DropdownItem
                        key={`${option.namespaceId}-${option.isPublic}`}
                        className="d-flex align-items-center"
                        onClick={() => onSelectShareOption(option)}
                        data-testid={`share-notebook-option-${option.namespaceName}-${option.isPublic}`}
                    >
                        <ShareOptionComponent {...option} isSourcegraphDotCom={isSourcegraphDotCom} />
                    </DropdownItem>
                ))}
            </DropdownMenu>
        </Dropdown>
    )
}
