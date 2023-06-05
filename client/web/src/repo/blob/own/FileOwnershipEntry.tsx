import React from 'react'

import { mdiEmail } from '@mdi/js'

import { TeamAvatar } from '@sourcegraph/shared/src/components/TeamAvatar'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { Button, ButtonLink, Icon, LinkOrSpan, Tooltip } from '@sourcegraph/wildcard'

import {
    AssignedOwnerFields,
    CodeownersFileEntryFields,
    OwnerFields,
    RecentContributorOwnershipSignalFields,
    RecentViewOwnershipSignalFields,
} from '../../../graphql-operations'
import { PersonLink } from '../../../person/PersonLink'

import { OwnershipBadge } from './OwnershipBadge'

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
                            {/* In case of unresolved, but guessed GitHub team, ID is 0 and we don't need to provide a link to the non-existing team. */}
                            <LinkOrSpan to={owner.external ? undefined : `/teams/${owner.name}`}>
                                {owner.teamDisplayName || owner.name}
                            </LinkOrSpan>
                        </>
                    )}
                </div>
            </td>
            <td className={containerStyles.expanding}>
                {reasons.map(reason => (
                    <OwnershipBadge key={reason.title} reason={reason} />
                ))}
            </td>
        </tr>
    )
}
