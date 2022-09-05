import React from 'react'

import classNames from 'classnames'
import { parseISO } from 'date-fns'

import { numberWithCommas } from '@sourcegraph/common'
import * as GQL from '@sourcegraph/shared/src/schema'

import { ExpirationDate } from '../../productSubscription/ExpirationDate'

/**
 * Displays a before/after comparison table of a change to the invoice item of a product
 * subscription.
 */
export const ProductSubscriptionBeforeAfterInvoiceItem: React.FunctionComponent<
    React.PropsWithChildren<{
        beforeInvoiceItem: GQL.IProductSubscriptionInvoiceItem
        afterInvoiceItem: GQL.IProductSubscriptionInvoiceItem
        className?: string
    }>
> = ({ beforeInvoiceItem, afterInvoiceItem, className = '' }) => {
    const planChanged = beforeInvoiceItem.plan.billingPlanID !== afterInvoiceItem.plan.billingPlanID
    const userCountChanged = beforeInvoiceItem.userCount !== afterInvoiceItem.userCount
    return !planChanged && !userCountChanged ? (
        <div className={classNames('text-muted', className)}>No changes to subscription.</div>
    ) : (
        <ul className={classNames('pl-3', className)}>
            {planChanged && (
                <li>
                    {afterInvoiceItem.plan.pricePerUserPerYear > beforeInvoiceItem.plan.pricePerUserPerYear
                        ? 'Upgrade'
                        : 'Downgrade'}{' '}
                    plan from {beforeInvoiceItem.plan.name} to {afterInvoiceItem.plan.name}.
                </li>
            )}
            {userCountChanged && (
                <li>
                    {afterInvoiceItem.userCount > beforeInvoiceItem.userCount ? 'Increase' : 'Decrease'} user count from{' '}
                    {numberWithCommas(beforeInvoiceItem.userCount)} to {numberWithCommas(afterInvoiceItem.userCount)}.
                </li>
            )}
            <li>
                Prorated for remainder of subscription (through{' '}
                <ExpirationDate showTime={false} date={parseISO(afterInvoiceItem.expiresAt)} lowercase={true} />
                ).
            </li>
        </ul>
    )
}
