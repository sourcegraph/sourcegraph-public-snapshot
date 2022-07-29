import React from 'react'

import { mdiOpenInNew } from '@mdi/js'
import { parseISO } from 'date-fns'
import format from 'date-fns/format'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import * as GQL from '@sourcegraph/shared/src/schema'
import { CardBody, Icon, Tooltip } from '@sourcegraph/wildcard'

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
                            <Tooltip content={format(parseISO(event.date), 'PPpp')}>
                                <span>{format(parseISO(event.date), 'yyyy-MM-dd')}</span>
                            </Tooltip>
                        </td>
                        <td className="w-100">
                            <LinkOrSpan to={event.url} target="_blank" rel="noopener noreferrer">
                                {event.title}
                                {event.url && <Icon aria-hidden={true} className="ml-1" svgPath={mdiOpenInNew} />}
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
