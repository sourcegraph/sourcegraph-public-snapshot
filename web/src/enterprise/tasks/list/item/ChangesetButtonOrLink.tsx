import React from 'react'
import { Redirect } from 'react-router'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CreateOrPreviewChangesetButton, CreateOrPreviewChangesetButtonProps } from './CreateOrPreviewChangesetButton'

const LOADING = 'loading' as const

export const PENDING_CREATION = 'pending-creation' as const

export type ChangesetButtonOrLinkExistingChangeset =
    | typeof LOADING
    | null
    | typeof PENDING_CREATION
    | Pick<GQL.IDiscussionThread, 'idWithoutKind' | 'status' | 'url'>

interface Props extends CreateOrPreviewChangesetButtonProps {
    existingChangeset: ChangesetButtonOrLinkExistingChangeset
}

/**
 * A button to create/preview a changeset if no changeset exists yet, or else a link to the existing
 * changeset.
 */
export const ChangesetButtonOrLink: React.FunctionComponent<Props> = ({
    existingChangeset,
    disabled,
    className = '',
    ...props
}) =>
    existingChangeset === LOADING ? (
        <span className={`text-muted ${className}`}>Determining changeset status...</span>
    ) : existingChangeset === null || existingChangeset === PENDING_CREATION ? (
        <CreateOrPreviewChangesetButton
            {...props}
            onClick={props.onClick}
            disabled={disabled || existingChangeset === PENDING_CREATION}
            className={className}
        />
    ) : existingChangeset.status === GQL.ThreadStatus.PREVIEW ? (
        <Redirect to={existingChangeset.url} push={true} />
    ) : (
        <Link className={`btn btn-secondary ${className}`} to={existingChangeset.url}>
            Changeset #{existingChangeset.idWithoutKind}
        </Link>
    )
