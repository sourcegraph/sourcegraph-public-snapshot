import { mdiFileDocumentOutline, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import { H2, Icon, Link, LoadingSpinner, Text } from '@sourcegraph/wildcard'

import { Client } from '../../api/client'
import { useApiCaller } from '../../api/hooks/useApiClient'
import type { Invoice } from '../../api/teamSubscriptions'

import { humanizeDate, usdCentsToHumanString } from './utils'

import styles from './InvoiceHistory.module.scss'

const invoicesCall = Client.getCurrentSubscriptionInvoices()

export const InvoiceHistory: React.FC = () => {
    const { loading, error, data } = useApiCaller(invoicesCall)

    if (loading) {
        return <LoadingSpinner />
    }

    if (error) {
        // TODO: handle error
        return null
    }

    if (!data) {
        // TODO: why empty response - handle it
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
    <li className={styles.invoice}>
        <div className={classNames('text-muted', styles.invoiceCol)}>
            <Icon aria-hidden={true} svgPath={mdiFileDocumentOutline} />
            <Text as="span">{invoice.periodEnd ? humanizeDate(invoice.periodEnd) : '(no date)'}</Text>
        </div>

        <div className={classNames('font-weight-medium', styles.invoiceCol)}>
            <Text as="span" className="text-muted">
                {usdCentsToHumanString(invoice.amountDue)}
            </Text>
            <Text as="span" className="text-capitalize">
                {invoice.status}
            </Text>
            {invoice.hostedInvoiceUrl ? (
                <Link to={invoice.hostedInvoiceUrl} target="_blank" rel="noopener" className={styles.invoiceLink}>
                    <Text as="span">Get Invoice</Text>
                    <Icon aria-hidden={true} svgPath={mdiOpenInNew} className={styles.invoiceLinkIcon} />
                </Link>
            ) : (
                '-'
            )}
        </div>
    </li>
)
