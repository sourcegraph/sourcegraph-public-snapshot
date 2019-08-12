import React, { useState } from 'react'
import { DropdownItem, DropdownMenu } from 'reactstrap'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { DropdownMenuFilter } from '../../../../components/dropdownMenuFilter/DropdownMenuFilter'
import { useThreads } from '../../../campaigns/detail/threads/useThreads'
import { ThreadStateIcon } from '../../common/threadState/ThreadStateIcon'

interface Props {
    /** Called when the user selects a thread in the menu. */
    onSelect: (thread: Pick<GQL.IThread, 'id'>) => void
}

const LOADING = 'loading' as const

/**
 * A dropdown menu with a list of threads.
 */
export const ThreadsDropdownMenu: React.FunctionComponent<Props> = ({ onSelect, ...props }) => {
    const [query, setQuery] = useState('')
    const threads = useThreads(true)
    return (
        <DropdownMenu {...props}>
            <DropdownMenuFilter value={query} onChange={setQuery} placeholder="Filter threads" header="Add a thread" />
            {threads === LOADING ? (
                <DropdownItem header={true} className="py-1">
                    Loading threads...
                </DropdownItem>
            ) : isErrorLike(threads) ? (
                <DropdownItem header={true} className="py-1">
                    Error loading threads
                </DropdownItem>
            ) : (
                threads.nodes
                    .filter(thread => thread.title.toLowerCase().includes(query.toLowerCase()))
                    .slice(0, 20 /* TODO!(sqs) hack */)
                    .map(thread => (
                        // tslint:disable-next-line: jsx-no-lambda
                        <DropdownItem key={thread.id} onClick={() => onSelect(thread)}>
                            <ThreadStateIcon thread={thread} className="small mr-1" />
                            <small className="text-muted">#{thread.number}</small> {thread.title}
                        </DropdownItem>
                    ))
            )}
        </DropdownMenu>
    )
}
