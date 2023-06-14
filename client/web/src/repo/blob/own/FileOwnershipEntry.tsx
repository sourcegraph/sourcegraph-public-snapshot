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
import { RemoveOwnerButton } from './RemoveOwnerButton'

import containerStyles from './OwnerList.module.scss'

interface Props {
    owner: OwnerFields
    reasons: OwnershipReason[]
    makeOwnerButton?: React.ReactElement
    repoID: string
    filePath: string
    setRemoveOwnerError: any
    refetch: any
}

type OwnershipReason =
    | CodeownersFileEntryFields
    | RecentContributorOwnershipSignalFields
    | RecentViewOwnershipSignalFields
    | AssignedOwnerFields

export const FileOwnershipEntry: React.FunctionComponent<Props> = ({
    owner,
    reasons,
    makeOwnerButton,
    repoID,
    filePath,
    refetch,
    setRemoveOwnerError,
}) => {
    const findEmail = (): string | undefined => {
        if (owner.__typename !== 'Person') {
            return undefined
        }
        if (owner.user?.primaryEmail) {
            return owner.user?.primaryEmail.email
        }
        return owner.email
    }

    const userID = owner.__typename === 'Person' && owner.user?.__typename === 'User' ? owner.user.id : undefined
    const teamID = owner.__typename === 'Team' ? owner.id : undefined

    const email = findEmail()

    const assignedOwnerReasons: AssignedOwnerFields[] = reasons
        .filter(value => value.__typename === 'AssignedOwner')
        .map(val => val as AssignedOwnerFields)
    const isDirectAssigned = assignedOwnerReasons.some(value => value.isDirectMatch)
    const hasAssigned = assignedOwnerReasons.length > 0

    return (
        <tr>
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
            <td className={containerStyles.fitting}>
                <span className={containerStyles.editButtonColumn}>
                    <div className={containerStyles.removeAddButton}>
                        {makeOwnerButton ||
                            (hasAssigned && (
                                <RemoveOwnerButton
                                    onSuccess={refetch}
                                    onError={setRemoveOwnerError}
                                    repoId={repoID}
                                    path={filePath}
                                    userID={userID}
                                    teamID={teamID}
                                    isDirectAssigned={isDirectAssigned}
                                />
                            ))}
                    </div>
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
                </span>
            </td>
        </tr>
    )
}
