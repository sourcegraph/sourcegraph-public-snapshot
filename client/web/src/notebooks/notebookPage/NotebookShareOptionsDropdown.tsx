import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronUpIcon from 'mdi-react/ChevronUpIcon'
import DomainIcon from 'mdi-react/DomainIcon'
import LockIcon from 'mdi-react/LockIcon'
import WebIcon from 'mdi-react/WebIcon'
import React, { useState, useCallback, useMemo } from 'react'
import { Dropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

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
    Omit<ShareOption, 'namespaceId'> & { isSourcegraphDotCom: boolean }
> = ({ isSourcegraphDotCom, namespaceType, namespaceName, isPublic }) => {
    if (namespaceType === 'User') {
        if (isPublic) {
            const PublicIcon = isSourcegraphDotCom ? WebIcon : DomainIcon
            const publicText = isSourcegraphDotCom ? 'Public' : 'Instance'
            return (
                <>
                    <PublicIcon className="mr-2" size="1.15rem" /> {publicText}
                </>
            )
        }
        return (
            <>
                <LockIcon className="mr-2" size="1.15rem" /> Private
            </>
        )
    }
    return (
        <>
            <OrgAvatar org={namespaceName} className="d-inline-flex mr-2" size="sm" /> {namespaceName}
        </>
    )
}

export const NotebookShareOptionsDropdown: React.FunctionComponent<NotebookShareOptionsDropdownProps> = ({
    isSourcegraphDotCom,
    telemetryService,
    authenticatedUser,
    selectedShareOption,
    onSelectShareOption,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleOpen = useCallback(() => {
        telemetryService.log('NotebookVisibilitySettingsDropdownToggled')
        setIsOpen(value => !value)
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
        <Dropdown isOpen={isOpen} toggle={toggleOpen}>
            <DropdownToggle
                className={styles.button}
                outline={true}
                data-testid="share-notebook-options-dropdown-toggle"
            >
                <span className="d-flex align-items-center">
                    <ShareOptionComponent {...selectedShareOption} isSourcegraphDotCom={isSourcegraphDotCom} />
                </span>
                <span className="ml-5">{isOpen ? <ChevronUpIcon /> : <ChevronDownIcon />}</span>
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
