import React from 'react'
import { DropdownItem, DropdownMenu } from 'reactstrap'
import * as GQL from '../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { useThreads } from '../campaigns/detail/threads/useThreads'

interface Props {
    /** Called when the user selects a thread in the menu. */
    onSelect: (thread: Pick<GQL.ThreadOrIssueOrChangeset, 'id'>) => void
}

const LOADING = 'loading' as const

/**
 * A dropdown menu with a list of threads.
 */
export const ThreadsDropdownMenu: React.FunctionComponent<Props> = ({ onSelect, ...props }) => {
    const threads = useThreads()
    return (
        <DropdownMenu {...props}>
            {threads === LOADING ? (
                <DropdownItem header={true} className="py-1">
                    Loading threads...
                </DropdownItem>
            ) : isErrorLike(threads) ? (
                <DropdownItem header={true} className="py-1">
                    Error loading threads
                </DropdownItem>
            ) : (
                threads.nodes.map(thread => (
                    // tslint:disable-next-line: jsx-no-lambda
                    <DropdownItem key={thread.id} onClick={() => onSelect(thread)}>
                        <small className="text-muted">#{thread.number}</small> {thread.title}
                    </DropdownItem>
                ))
            )}
        </DropdownMenu>
    )
}
