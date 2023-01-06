import React, { useCallback, useState } from 'react'

import { mdiChat, mdiChevronDown, mdiChevronUp, mdiEmail } from '@mdi/js'
import { AccordionButton, AccordionItem, AccordionPanel } from '@reach/accordion'

import { Badge, Button, Icon } from '@sourcegraph/wildcard'

interface Props {
    handle: string
    email: string
    reasons: OwnershipReasonDetails[]
}

interface OwnershipReasonDetails {
    title: string
    description: string
}

export const FileOwnershipReasons: React.FunctionComponent<Props> = ({ handle, email, reasons }) => {
    const [isExpanded, setIsExpanded] = useState<boolean>(false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(() => {
        setIsExpanded(!isExpanded)
    }, [isExpanded])
    return (
        <AccordionItem as={React.Fragment}>
            <tr>
                <td>
                    <div className="d-flex">
                        <AccordionButton
                            as={Button}
                            variant="icon"
                            className="d-none d-sm-block mx-1"
                            aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                            onClick={toggleIsExpanded}
                        >
                            <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} />
                        </AccordionButton>
                        <Button variant="icon" className="mr-2">
                            <Icon svgPath={mdiEmail} aria-label="email" />
                        </Button>
                        <Button variant="icon">
                            <Icon svgPath={mdiChat} aria-label="chat" />
                        </Button>
                    </div>
                </td>
                <td>{handle}</td>
                <td>{email}</td>
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
