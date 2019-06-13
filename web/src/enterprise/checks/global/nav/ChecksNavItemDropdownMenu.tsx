import AddIcon from 'mdi-react/AddIcon'
import TableIcon from 'mdi-react/TableIcon'
import React, { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { DropdownItem, DropdownMenu } from 'reactstrap'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { fetchDiscussionThreads } from '../../../../discussions/backend'

interface Props {
    className?: string
}

const LOADING: 'loading' = 'loading'

/**
 * A dropdown menu with a list of navigation links related to checks.
 */
export const ChecksNavItemDropdownMenu: React.FunctionComponent<Props> = ({ className = '' }) => {
    const [threadsOrError, setChecksOrError] = useState<typeof LOADING | GQL.IDiscussionThreadConnection | ErrorLike>(
        LOADING
    )

    // tslint:disable-next-line: no-floating-promises
    useMemo(async () => {
        try {
            setChecksOrError(await fetchDiscussionThreads({}).toPromise())
        } catch (err) {
            setChecksOrError(asError(err))
        }
    }, [])

    const MAX_CHECKS = 5 // TODO!(sqs): hack

    return (
        <DropdownMenu right={true} className={className}>
            <Link to="/checks/dashboard" className="dropdown-item d-flex align-items-center">
                <TableIcon className="icon-inline mr-1" /> Dashboard
            </Link>
            <DropdownItem divider={true} />
            {threadsOrError === LOADING ? (
                <DropdownItem header={true} className="py-1">
                    Loading checks...
                </DropdownItem>
            ) : (
                !isErrorLike(threadsOrError) && (
                    <>
                        <DropdownItem header={true} className="py-1">
                            Recent checks
                        </DropdownItem>
                        {threadsOrError.nodes.slice(0, MAX_CHECKS).map(thread => (
                            <Link key={thread.id} to={thread.url} className="dropdown-item text-truncate">
                                <small className="text-muted">#{thread.idWithoutKind}</small> {thread.title}
                            </Link>
                        ))}
                    </>
                )
            )}
        </DropdownMenu>
    )
}
