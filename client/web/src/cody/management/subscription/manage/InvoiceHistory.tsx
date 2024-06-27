import { mdiFileDocumentOutline, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { logger } from '@sourcegraph/common'
import { H2, Icon, Link, LoadingSpinner, Text } from '@sourcegraph/wildcard'

import { useSubscriptionInvoices } from '../../api/react-query/subscriptions'
import type { Invoice } from '../../api/teamSubscriptions'

import { humanizeDate, usdCentsToHumanString } from './utils'

import styles from './InvoiceHistory.module.scss'

export const InvoiceHistory: React.FC = () => {
    const { isLoading, isError, error, data } = useSubscriptionInvoices()

    if (isLoading) {
        return <LoadingSpinner />
    }

    if (isError) {
        logger.error('Error fetching current subscription invoices', error)
        return null
    }

    if (!data) {
        logger.error('Current subscription invoices are not available.')
        return null
    }

    return (
        <>
            <H2 className="mb-4">Invoice history</H2>

            <hr className={classNames('w-100', styles.divider)} />

            {data.invoices.length ? (
                <ul className="mb-0 list-unstyled">
                    {data.invoices.map(invoice => (
                        <InvoiceItem key={invoice.periodStart} invoice={invoice} />
                    ))}
                </ul>
            ) : (
                <Text>You have no invoices.</Text>
            )}
        </>
    )
}

const InvoiceItem: React.FC<{ invoice: Invoice }> = ({ invoice }) => (
    <li className="mt-3 d-flex justify-content-between align-items-center">
        <div className={classNames('d-flex align-items-center text-muted', styles.label)}>
            <Icon aria-hidden={true} svgPath={mdiFileDocumentOutline} />
            <Text as="span">{invoice.periodEnd ? humanizeDate(invoice.periodEnd) : '(no date)'}</Text>
        </div>

        <div className={classNames('d-flex align-items-center font-weight-medium', styles.price)}>
            <Text as="span" className="text-muted">
                {usdCentsToHumanString(invoice.amountDue)}
            </Text>
            <Text as="span" className="text-capitalize">
                {invoice.status}
            </Text>
            {invoice.hostedInvoiceUrl ? (
                <Link
                    to={invoice.hostedInvoiceUrl}
                    target="_blank"
                    rel="noopener"
                    className="d-flex align-items-center"
                >
                    <Text as="span">Get Invoice</Text>
                    <Icon aria-hidden={true} svgPath={mdiOpenInNew} className={styles.icon} />
                </Link>
            ) : (
                '-'
            )}
        </div>
    </li>
)
