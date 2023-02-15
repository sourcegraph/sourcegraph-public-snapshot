import React, { useState, useCallback } from 'react'

import addDays from 'date-fns/addDays'
import endOfDay from 'date-fns/endOfDay'
import { Observable } from 'rxjs'
import { catchError, map, startWith, switchMap, tap } from 'rxjs/operators'

import { asError, createAggregateError, isErrorLike } from '@sourcegraph/common'
import { gql, useMutation } from '@sourcegraph/http-client'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import {
    Alert,
    Button,
    useEventObservable,
    Link,
    Label,
    Input,
    ErrorAlert,
    Form,
    Select,
    useDebounce,
} from '@sourcegraph/wildcard'

import {
    GenerateProductLicenseForSubscriptionResult,
    GenerateProductLicenseForSubscriptionVariables,
} from '../../../../graphql-operations'
import { ExpirationDate } from '../../../productSubscription/ExpirationDate'
import { hasUnknownTags, ProductLicenseTags, UnknownTagWarning } from '../../../productSubscription/ProductLicenseTags'
import { debounce, DebouncedFunc } from 'lodash'

interface Props {
    subscriptionID: Scalars['ID']
    subscriptionAccount: string
    onGenerate: () => void
}

interface FormData {
    /** Comma-separated license tags. */
    tags: string
    customer: string
    plan: string
    userCount: number
    validDays: number
    expiresAt: number
}

const getEmptyFormData = (account: string): FormData => ({
    tags: '',
    customer: account,
    plan: 'enterprise-1',
    userCount: 1,
    validDays: 1,
    expiresAt: addDaysAndRoundToEndOfDay(1),
})

const DURATION_LINKS = [
    { label: '7 days', days: 7 },
    { label: '14 days', days: 14 },
    { label: '30 days', days: 30 },
    { label: '60 days', days: 60 },
    { label: '1 year', days: 366 }, // 366 not 365 to account for leap year
]

const GENERATE_PRODUCT_LICENSE = gql`
    mutation GenerateProductLicenseForSubscription($productSubscriptionID: ID!, $license: ProductLicenseInput!) {
        dotcom {
            generateProductLicenseForSubscription(productSubscriptionID: $productSubscriptionID, license: $license) {
                id
            }
        }
    }
`

const tagsFromString = (tagString: string): string[] => {
    return tagString
        .split(',')
        .map(item => item.trim())
        .filter(tag => tag !== '')
}

const getTagsFromFormData = (formData: FormData): string[] => {
    return [`customer:${formData.customer}`, `plan:${formData.plan}`, ...tagsFromString(formData.tags)]
}

/**
 * Displays a form to generate a new product license for a product subscription.
 *
 * For use on Sourcegraph.com by Sourcegraph teammates only.
 */
export const SiteAdminGenerateProductLicenseForSubscriptionForm: React.FunctionComponent<
    React.PropsWithChildren<Props>
> = ({ subscriptionID, subscriptionAccount, onGenerate }) => {
    const [formData, setFormData] = useState<FormData>(getEmptyFormData(subscriptionAccount))

    const onPlanChange = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(
        event => setFormData(formData => ({ ...formData, plan: event.target.value })),
        []
    )

    const onCustomerChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => setFormData(formData => ({ ...formData, customer: event.target.value })),
        []
    )

    const onTagsChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => setFormData(formData => ({ ...formData, tags: event.target.value || '' })),
        []
    )

    const onUserCountChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => setFormData(formData => ({ ...formData, userCount: event.target.valueAsNumber })),
        []
    )

    const setValidDays = useCallback((validDays: number): void => {
        setFormData(formData => ({
            ...formData,
            validDays,
            expiresAt: addDaysAndRoundToEndOfDay(validDays),
        }))
    }, [])
    const onValidDaysChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => setValidDays(Number.isNaN(event.currentTarget.valueAsNumber) ? 0 : event.currentTarget.valueAsNumber),
        [setValidDays]
    )

    const dismissAlert = useCallback((): void => setFormData(getEmptyFormData(subscriptionAccount)), [])

    const [generateLicense, { loading, error, data }] = useMutation<
        GenerateProductLicenseForSubscriptionResult['dotcom']['generateProductLicenseForSubscription'],
        GenerateProductLicenseForSubscriptionVariables
    >(GENERATE_PRODUCT_LICENSE, {
        variables: {
            productSubscriptionID: subscriptionID,
            license: {
                tags: getTagsFromFormData(formData),
                userCount: formData.userCount,
                expiresAt: Math.ceil(formData.expiresAt / 1000),
            },
        },
        onCompleted: onGenerate,
    })

    const plans =
        window.context.licenseInfo?.knownLicenseTags?.filter(tag => tag.startsWith('plan:')).map(tag => tag.slice(5)) ||
        []

    const onSubmit = useCallback<React.FormEventHandler>(
        event => {
            event.preventDefault()
            generateLicense()
        },
        [generateLicense]
    )

    const disableForm = loading || error
    const tags = useDebounce<string[]>(tagsFromString(formData.tags), 300)

    return (
        <div className="site-admin-generate-product-license-for-subscription-form">
            {data ? (
                <div className="border rounded border-success mb-5">
                    <Alert
                        variant="success"
                        className="border-top-0 border-left-0 border-right-0 rounded-0 mb-0 d-flex align-items-center justify-content-between px-3 py-2"
                    >
                        <span>Generated product license.</span>
                        <Button onClick={dismissAlert} autoFocus={true} variant="primary">
                            Dismiss
                        </Button>
                    </Alert>
                </div>
            ) : (
                <Form onSubmit={onSubmit}>
                    <Alert variant="info">
                        Please read the{' '}
                        <Link to="https://handbook.sourcegraph.com/ce/license_keys#how-to-create-a-license-key-for-a-new-prospect-or-new-customer">
                            guide for how to create a license key for a new prospect or new customer.
                        </Link>
                    </Alert>
                    <div className="form-group">
                        <Label htmlFor="site-admin-create-product-subscription-page__customer_input">Customer</Label>
                        <Input
                            id="site-admin-create-product-subscription-page__customer_input"
                            type="text"
                            disabled={disableForm}
                            value={formData.customer || ''}
                            onChange={onCustomerChange}
                        />
                        <small className="form-text text-muted">
                            Name of the customer. Defaults to the Account name from the subscription.
                        </small>
                    </div>
                    <div className="form-group">
                        <Select
                            id="site-admin-create-product-subscription-page__plan_select"
                            label="Plan"
                            disabled={disableForm}
                            value={formData.plan}
                            onChange={onPlanChange}
                            className="mb-0"
                        >
                            {plans.map(plan => (
                                <option key={plan} value={plan}>
                                    {plan}
                                </option>
                            ))}
                        </Select>
                        <small className="form-text text-muted">Subscription plan. Required parameter.</small>
                    </div>
                    <div className="form-group">
                        <Label htmlFor="site-admin-create-product-subscription-page__tags">Tags</Label>
                        <Input
                            type="text"
                            id="site-admin-create-product-subscription-page__tags"
                            disabled={disableForm}
                            value={formData.tags}
                            onChange={onTagsChange}
                            list="known-tags"
                            className="mb-0"
                        />
                        <datalist id="known-tags">
                            {tags.map(tag => (
                                <option key={tag} value={tag}>
                                    {tag}
                                </option>
                            ))}
                        </datalist>
                        <div className="mt-1">
                            <ProductLicenseTags tags={tags} />
                        </div>
                        {hasUnknownTags(tags) && <UnknownTagWarning />}
                        <small className="form-text text-muted">
                            Comma separated list of tags. Tags restrict a license.
                        </small>
                    </div>
                    <div className="form-group">
                        <Label htmlFor="site-admin-create-product-subscription-page__userCount">Users</Label>
                        <Input
                            type="number"
                            min={1}
                            id="site-admin-create-product-subscription-page__userCount"
                            disabled={disableForm}
                            value={formData.userCount || ''}
                            onChange={onUserCountChange}
                        />
                    </div>
                    <div className="form-group">
                        <Label htmlFor="site-admin-create-product-subscription-page__validDays">Valid for (days)</Label>
                        <Input
                            type="number"
                            id="site-admin-create-product-subscription-page__validDays"
                            disabled={disableForm}
                            value={formData.validDays || ''}
                            min={1}
                            max={2000} // avoid overflowing int32
                            onChange={onValidDaysChange}
                        />
                        <small className="form-text text-muted">
                            {formData.expiresAt !== null ? (
                                <ExpirationDate
                                    date={formData.expiresAt}
                                    showTime={true}
                                    showRelative={true}
                                    showPrefix={true}
                                />
                            ) : (
                                <>&nbsp;</>
                            )}
                        </small>
                        <small className="form-text text-muted d-block mt-1">
                            Set to{' '}
                            {DURATION_LINKS.map(({ label, days }) => (
                                <Button
                                    key={days}
                                    className="p-0 mr-2"
                                    onClick={() => setValidDays(days)}
                                    variant="link"
                                    size="sm"
                                >
                                    {label}
                                </Button>
                            ))}
                        </small>
                    </div>
                    <Button type="submit" disabled={disableForm} variant={disableForm ? 'secondary' : 'primary'}>
                        Generate license
                    </Button>
                </Form>
            )}
            {error && <ErrorAlert className="mt-3" error={error} />}
        </div>
    )
}

/**
 * Adds 1 day to the current date, then rounds it up to midnight in the client's timezone. This is a
 * generous interpretation of "valid for N days" to avoid confusion over timezones or "will it
 * expire at the beginning of the day or at the end of the day?"
 */
function addDaysAndRoundToEndOfDay(amount: number): number {
    return endOfDay(addDays(Date.now(), amount)).getTime()
}
