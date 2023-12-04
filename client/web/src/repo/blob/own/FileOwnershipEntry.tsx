import React from 'react'

import { mdiEmail } from '@mdi/js'
import { useNavigate } from 'react-router-dom'

import { TeamAvatar } from '@sourcegraph/shared/src/components/TeamAvatar'
import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { Button, ButtonLink, Icon, LinkOrSpan, Tooltip } from '@sourcegraph/wildcard'

import type {
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
    isDirectory: boolean
    refetch: any
    canRemoveOwner?: boolean
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
    setRemoveOwnerError,
    isDirectory,
    refetch,
    canRemoveOwner,
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

    const sortReasons = () => (reason1: OwnershipReason, reason2: OwnershipReason) =>
        getOwnershipReasonPriority(reason2) - getOwnershipReasonPriority(reason1)

    const navigate = useNavigate()
    const refresh = (): Promise<any> => {
        if (!isDirectory) {
            return Promise.resolve(navigate(0))
        }
        return Promise.resolve(refetch())
    }

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
                {reasons
                    .slice()
                    .sort(sortReasons())
                    .map(reason => (
                        <OwnershipBadge key={reason.title} reason={reason} />
                    ))}
            </td>
            <td className={containerStyles.fitting}>
                <span className={containerStyles.editButtonColumn}>
                    <div className={containerStyles.removeAddButton}>
                        {makeOwnerButton ||
                            (canRemoveOwner && hasAssigned && (
                                <RemoveOwnerButton
                                    onSuccess={refresh}
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

const getOwnershipReasonPriority = (reason: OwnershipReason): number => {
    switch (reason.__typename ?? '') {
        case 'CodeownersFileEntry': {
            return 4
        }
        case 'AssignedOwner': {
            return 3
        }
        case 'RecentContributorOwnershipSignal': {
            return 2
        }
        case 'RecentViewOwnershipSignal': {
            return 1
        }
        default: {
            return 0
        }
    }
}
