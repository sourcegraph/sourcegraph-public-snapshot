import React, { useCallback, useState } from 'react'

import { mdiChevronDown, mdiChevronUp, mdiEmail } from '@mdi/js'
import { AccordionButton, AccordionItem, AccordionPanel } from '@reach/accordion'

import { Badge, Button, ButtonLink, Icon } from '@sourcegraph/wildcard'

import { CodeownersFileEntryFields, OwnerFields } from '../../../graphql-operations'
import { PersonLink } from '../../../person/PersonLink'
import { UserAvatar } from '../../../user/UserAvatar'

import styles from './FileOwnershipEntry.module.scss'

interface Props {
    person: OwnerFields
    reasons: CodeownersFileEntryFields[]
}

export const FileOwnershipEntry: React.FunctionComponent<Props> = ({ person, reasons }) => {
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
                        <ButtonLink
                            variant="icon"
                            className="mr-2"
                            disabled={!person.email}
                            to={person.email ? `mailto:${person.email}` : undefined}
                        >
                            <Icon svgPath={mdiEmail} aria-label="email" />
                        </ButtonLink>
                    </div>
                </td>
                <td>
                    <UserAvatar user={person} className="mx-2" />
                    <PersonLink person={person} />
                </td>
                <td>
                    {reasons.map(reason => (
                        <Badge key={reason.title} tooltip={reason.description} className={styles.badge}>
                            {reason.title}
                        </Badge>
                    ))}
                </td>
            </tr>
            <AccordionPanel as="tr">
                <td colSpan={4}>
                    <ul className={styles.reasons}>
                        {reasons.map(reason => (
                            <li key={reason.title}>
                                <Badge className={styles.badge}>{reason.title}</Badge> {reason.description}
                            </li>
                        ))}
                    </ul>
                </td>
            </AccordionPanel>
        </AccordionItem>
    )
}
