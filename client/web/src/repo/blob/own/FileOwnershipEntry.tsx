import React, { useCallback, useState } from 'react'

import { mdiChevronDown, mdiChevronUp, mdiEmail } from '@mdi/js'
import { AccordionButton, AccordionItem, AccordionPanel } from '@reach/accordion'

import { TeamAvatar } from '@sourcegraph/shared/src/components/TeamAvatar'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { Badge, Button, ButtonLink, Icon, Link, Tooltip } from '@sourcegraph/wildcard'

import { CodeownersFileEntryFields, OwnerFields } from '../../../graphql-operations'
import { PersonLink } from '../../../person/PersonLink'

import styles from './FileOwnershipEntry.module.scss'

interface Props {
    owner: OwnerFields
    reasons: CodeownersFileEntryFields[]
}

export const FileOwnershipEntry: React.FunctionComponent<Props> = ({ owner, reasons }) => {
    const [isExpanded, setIsExpanded] = useState<boolean>(false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(() => {
        setIsExpanded(!isExpanded)
    }, [isExpanded])

    const email = owner.__typename === 'Person' ? owner.email : undefined

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
                        <Tooltip content={email ? `Email ${email}` : 'No email address'} placement="top">
                            {email ? (
                                <ButtonLink variant="icon" to={`mailto:${email}`}>
                                    <Icon svgPath={mdiEmail} aria-label="email" />
                                </ButtonLink>
                            ) : (
                                <Button variant="icon" disabled={true}>
                                    <Icon svgPath={mdiEmail} aria-label="email" />
                                </Button>
                            )}
                        </Tooltip>
                    </div>
                </td>
                <td>
                    <div className="d-flex align-items-center mr-2">
                        {owner.__typename === 'Person' && (
                            <>
                                <UserAvatar user={owner} className="mx-2" inline={true} />
                                <PersonLink person={owner} />
                            </>
                        )}
                        {owner.__typename === 'Team' && (
                            <>
                                <TeamAvatar
                                    team={{ ...owner, displayName: owner.teamDisplayName }}
                                    className="mx-2"
                                    inline={true}
                                />
                                <Link to={`/teams/${owner.name}`}>{owner.teamDisplayName || owner.name}</Link>
                            </>
                        )}
                    </div>
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
