import React from 'react'

import { parseISO } from 'date-fns'
import format from 'date-fns/format'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import * as GQL from '@sourcegraph/shared/src/schema'
import { CardBody, Icon } from '@sourcegraph/wildcard'

export const ProductSubscriptionHistory: React.FunctionComponent<
    React.PropsWithChildren<{
        productSubscription: Pick<GQL.IProductSubscription, 'events'>
    }>
> = ({ productSubscription }) =>
    productSubscription.events.length > 0 ? (
        <table className="table mb-0">
            <thead>
                <tr>
                    <th>Date</th>
                    <th>Description</th>
                </tr>
            </thead>
            <tbody>
                {productSubscription.events.map(event => (
                    <tr key={event.id}>
                        <td className="text-nowrap">
                            <span data-tooltip={format(parseISO(event.date), 'PPpp')}>
                                {format(parseISO(event.date), 'yyyy-MM-dd')}
                            </span>
                        </td>
                        <td className="w-100">
                            <LinkOrSpan to={event.url} target="_blank" rel="noopener noreferrer">
                                {event.title}
                                {event.url && <Icon className="ml-1" as={ExternalLinkIcon} />}
                            </LinkOrSpan>
                            {event.description && <small className="d-block text-muted">{event.description}</small>}
                        </td>
                    </tr>
                ))}
            </tbody>
        </table>
    ) : (
        <CardBody>
            <span className="text-muted">No events</span>
        </CardBody>
    )
