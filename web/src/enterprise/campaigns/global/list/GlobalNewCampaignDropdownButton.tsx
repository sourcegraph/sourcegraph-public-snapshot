import PlusBoxIcon from 'mdi-react/PlusBoxIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'

const queryNamespaces = (): Observable<GQL.Namespace[]> =>
    queryGraphQL(
        gql`
            query ViewerNamespaces {
                viewerNamespaces {
                    __typename
                    id
                    namespaceName
                    url
                }
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.viewerNamespaces)
    )

interface Props {
    className?: string
}

export const GlobalNewCampaignDropdownButton: React.FunctionComponent<Props> = ({ className = '' }) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const [namespaces, setNamespaces] = useState<GQL.Namespace[] | ErrorLike>()
    useEffect(() => {
        const subscription = queryNamespaces().subscribe(setNamespaces, err => setNamespaces(asError(err)))
        return () => subscription.unsubscribe()
    }, [])

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className={className}>
            <DropdownToggle color="" className="btn btn-primary" caret={true}>
                <PlusBoxIcon className="icon-inline mr-2" /> New campaign
            </DropdownToggle>
            <DropdownMenu>
                <DropdownItem header={true}>New campaign in...</DropdownItem>
                <DropdownItem divider={true} />
                {namespaces === undefined ? (
                    <DropdownItem header={true} className="py-1">
                        Loading...
                    </DropdownItem>
                ) : isErrorLike(namespaces) ? (
                    <DropdownItem header={true} className="py-1">
                        Error loading namespaces
                    </DropdownItem>
                ) : (
                    namespaces.map(({ id, namespaceName, url }) => (
                        <Link key={id} className="dropdown-item" to={`${url}/campaigns/new`}>
                            {namespaceName}
                        </Link>
                    ))
                )}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
