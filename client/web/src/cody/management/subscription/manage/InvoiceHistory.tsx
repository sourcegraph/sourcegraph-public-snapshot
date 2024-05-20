import { Link, LoadingSpinner } from '@sourcegraph/wildcard'

import { Client } from '../../api/client'
import { useApiCaller } from '../../api/hooks/useApiClient'

import { humanizeDate, usdCentsToHumanString } from './utils'

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
        <div className="block p-6 bg-white border border-separator-gray rounded-md drop-shadow-sm mb-8">
            <h2 className="mb-4">Invoice history</h2>
            <hr className="mb-4" />
            {data.invoices.length ? (
                <ul>
                    {data.invoices.map(invoice => {
                        return (
                            <div key={invoice.periodStart} className="flex flex-row space-x-4">
                                <span className="flex-grow text-muted inline-block">
                                    📄 {invoice.periodEnd ? humanizeDate(invoice.periodEnd) : '(no date)'}
                                </span>
                                <span className="text-muted">{usdCentsToHumanString(invoice.amountDue)}</span>
                                <span>
                                    {invoice.status.charAt(0).toUpperCase() + invoice.status.slice(1).toLowerCase()}
                                </span>
                                {invoice.hostedInvoiceUrl ? (
                                    <Link to={invoice.hostedInvoiceUrl} target="_blank">
                                        Get invoice
                                    </Link>
                                ) : (
                                    '-'
                                )}
                            </div>
                        )
                    })}
                </ul>
            ) : (
                <p>You have no invoices.</p>
            )}
        </div>
    )
}
