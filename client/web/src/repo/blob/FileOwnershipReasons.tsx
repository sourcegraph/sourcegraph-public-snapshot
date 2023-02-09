import React, { useCallback, useState } from 'react'

import { mdiChat, mdiChevronDown, mdiChevronUp, mdiEmail } from '@mdi/js'
import { AccordionButton, AccordionItem, AccordionPanel } from '@reach/accordion'

import { Maybe } from '@sourcegraph/shared/src/graphql-operations'
import { Badge, Button, Icon } from '@sourcegraph/wildcard'

import { PersonLink } from '../../person/PersonLink'
import { UserAvatar } from '../../user/UserAvatar'

interface OwnerPerson {
    email: string
    displayName: string
    avatarURL: Maybe<string>
    user: Maybe<{
        username: string
        displayName: Maybe<string>
        url: string
    }>
}

interface Props {
    person: OwnerPerson
    reasons: OwnershipReasonDetails[]
}

interface OwnershipReasonDetails {
    title: string
    description: string
}

export const FileOwnershipReasons: React.FunctionComponent<Props> = ({ person, reasons }) => {
    const [isExpanded, setIsExpanded] = useState<boolean>(false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(() => {
        setIsExpanded(!isExpanded)
    }, [isExpanded])
    return (
        <AccordionItem as="tbody">
            <tr>
                <td>
                    <AccordionButton
                        as={Button}
                        variant="icon"
                        className="d-none d-sm-block mr-2"
                        aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                        onClick={toggleIsExpanded}
                    >
                        <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} />
                    </AccordionButton>
                </td>
                <td>
                    <div className="d-flex">
                        <Button variant="icon" className="mr-2">
                            <Icon svgPath={mdiEmail} aria-label="email" />
                        </Button>
                        <Button variant="icon">
                            <Icon svgPath={mdiChat} aria-label="chat" />
                        </Button>
                    </div>
                </td>
                <td>
                    <UserAvatar user={person} className="mx-2" />
                    <PersonLink person={person} />
                </td>
                <td>
                    {reasons.map(reason => (
                        <Badge key={reason.title} tooltip={reason.description}>
                            {reason.title}
                        </Badge>
                    ))}
                </td>
            </tr>
            <AccordionPanel as="tr">
                <td colSpan={4}>
                    <ul>
                        {reasons.map(reason => (
                            <li key={reason.title}>
                                <Badge>{reason.title}</Badge> {reason.description}
                            </li>
                        ))}
                    </ul>
                </td>
            </AccordionPanel>
        </AccordionItem>
    )
}
