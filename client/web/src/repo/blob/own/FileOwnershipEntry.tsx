import React, { useCallback, useState } from 'react'

import { mdiEmail } from '@mdi/js'

import { TeamAvatar } from '@sourcegraph/shared/src/components/TeamAvatar'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { Badge, Button, ButtonLink, Icon, Link, Tooltip } from '@sourcegraph/wildcard'

import {
    AssignedOwnerFields,
    CodeownersFileEntryFields,
    OwnerFields,
    RecentContributorOwnershipSignalFields,
    RecentViewOwnershipSignalFields,
} from '../../../graphql-operations'
import { PersonLink } from '../../../person/PersonLink'

import styles from './FileOwnershipEntry.module.scss'
import containerStyles from './FileOwnershipPanel.module.scss'

interface Props {
    owner: OwnerFields
    reasons: OwnershipReason[]
}

type OwnershipReason =
    | CodeownersFileEntryFields
    | RecentContributorOwnershipSignalFields
    | RecentViewOwnershipSignalFields
    | AssignedOwnerFields

export const FileOwnershipEntry: React.FunctionComponent<Props> = ({ owner, reasons }) => {
    // TODO remove callback and state
    const [isExpanded, setIsExpanded] = useState<boolean>(false)
    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(() => {
        setIsExpanded(!isExpanded)
    }, [isExpanded])

    const findEmail = (): string | undefined => {
        if (owner.__typename !== 'Person') {
            return undefined
        }
        if (owner.user?.primaryEmail) {
            return owner.user?.primaryEmail.email
        }
        return owner.email
    }

    const email = findEmail()

    return (
        <tr>
            <td className={containerStyles.fitting}>
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
            <td className={`${containerStyles.fitting} ${containerStyles.moreSpace}`}>
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
            <td className={containerStyles.expanding}>
                {reasons.map(reason => {
                    // For CODEOWNERS, the Badge links to the codeowners rule
                    let link: string | null = null
                    if (reason.__typename === 'CodeownersFileEntry') {
                        link = `${reason.codeownersFile.url}?L${reason.ruleLineMatch}`
                    }
                    return link !== null ? (
                        <Badge
                            key={reason.title}
                            tooltip={reason.description}
                            className={styles.badge}
                            as={Link}
                            to={link}
                        >
                            {reason.title}
                        </Badge>
                    ) : (
                        <Badge key={reason.title} tooltip={reason.description} className={styles.badge}>
                            {reason.title}
                        </Badge>
                    )
                })}
            </td>
        </tr>
    )
}
