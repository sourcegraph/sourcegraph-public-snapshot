import React from 'react'
import { Link } from 'react-router-dom'
import { RepositoryIcon } from '../../../../../shared/src/components/icons'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ErrorLike } from '../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../shared/src/util/strings'
import { Form } from '../../../components/Form'
import { TasksIcon } from '../../tasks/icons'
import { ThreadSettings } from '../../threads/settings'
import { ActionsIcon, DiffIcon, GitCommitIcon } from '../icons'

interface Props {
    thread: GQL.IDiscussionThread
    threadSettings: ThreadSettings

    className?: string
}

/**
 * A form to create a non-preview changeset from a preview changeset.
 */
export const CreateChangesetFromPreviewForm: React.FunctionComponent<Props> = ({ className = '', ...props }) => (
    <Form className={` ${className}`}>hello TODO!(sqs) add title and description form</Form>
)
