import React from 'react'
import { upperFirst } from 'lodash'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { AbstractThreadListItem } from '../list/AbstractThreadListItem'
import { ThreadStateIcon } from '../common/threadState/ThreadStateIcon'
import { Link } from 'react-router-dom'
import { displayRepoName } from '../../../../../shared/src/components/RepoFileLink'
import CloseIcon from 'mdi-react/CloseIcon'

interface Props {
    threadUpdatePreview: GQL.IThreadUpdatePreview

    showRepository?: boolean

    className?: string
}

const OPERATION_VERB: Record<GQL.ThreadUpdateOperation, string> = {
    [GQL.ThreadUpdateOperation.CREATION]: 'Create',
    [GQL.ThreadUpdateOperation.UPDATE]: 'Update',
    [GQL.ThreadUpdateOperation.DELETION]: 'Close',
}

/**
 * An item in the list of thread update previews.
 *
 * TODO!(sqs): handle more kinds of changes
 */
export const ThreadUpdatePreviewListItem: React.FunctionComponent<Props> = ({
    threadUpdatePreview: preview,
    showRepository = true,
    className,
}) => (
    <AbstractThreadListItem
        left={
            preview.newThread ? (
                <ThreadStateIcon thread={preview.newThread} />
            ) : (
                <CloseIcon className="icon-inline text-danger" />
            )
        }
        title={
            preview.newTitle !== null ? (
                <span>
                    {preview.newTitle} <s className="small font-weight-normal text-muted">{preview.oldTitle}</s>
                </span>
            ) : (
                (preview.oldThread || preview.newThread)!.title
            )
        }
        detail={[
            <span key={0} className="text-muted mr-2">
                {OPERATION_VERB[preview.operation]} {preview.newThread && preview.newThread.isDraft ? 'draft' : ''}{' '}
                {(preview.newThread || preview.oldThread)!.kind.toLowerCase()}{' '}
                {showRepository && (
                    <>
                        in{' '}
                        <Link to={(preview.newThread || preview.oldThread)!.repository.url}>
                            {displayRepoName((preview.newThread || preview.oldThread)!.repository.name)}
                        </Link>
                    </>
                )}
            </span>,
        ]}
        className={className}
    />
)
