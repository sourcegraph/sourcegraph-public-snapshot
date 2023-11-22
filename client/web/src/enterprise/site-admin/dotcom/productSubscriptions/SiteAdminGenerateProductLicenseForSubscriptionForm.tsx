import React, { useState, useCallback } from 'react'

import addDays from 'date-fns/addDays'
import endOfDay from 'date-fns/endOfDay'

import { useMutation } from '@sourcegraph/http-client'
import type { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import {
    Alert,
    Button,
    Link,
    Input,
    ErrorAlert,
    Form,
    Select,
    useDebounce,
    H3,
    Modal,
    Text,
} from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'
import type {
    GenerateProductLicenseForSubscriptionResult,
    GenerateProductLicenseForSubscriptionVariables,
    ProductLicenseFields,
    ProductLicenseInfoFields,
} from '../../../../graphql-operations'
import { ExpirationDate } from '../../../productSubscription/ExpirationDate'
import { hasUnknownTags, ProductLicenseTags, UnknownTagWarning } from '../../../productSubscription/ProductLicenseTags'

import { GENERATE_PRODUCT_LICENSE } from './backend'

interface License extends Omit<ProductLicenseFields, 'subscription' | 'info'> {
    info: Omit<ProductLicenseInfoFields, 'productNameWithBrand'> | null
}
interface Props {
    subscriptionID: Scalars['ID']
    subscriptionAccount: string
    latestLicense: License | undefined
    onGenerate: () => void
    onCancel: () => void
}

interface FormData {
    /** Comma-separated license tags. */
    tags: string
    customer: string
    salesforceSubscriptionID: string
    salesforceOpportunityID: string
    plan: string
    userCount: number
    expiresAt: Date
}

const getEmptyFormData = (account: string, latestLicense: License | undefined): FormData => ({
    tags: '',
    customer: account,
    salesforceSubscriptionID: latestLicense?.info?.salesforceSubscriptionID ?? '',
    salesforceOpportunityID: latestLicense?.info?.salesforceOpportunityID ?? '',
    plan: 'enterprise-1',
    userCount: 1,
    expiresAt: endOfDay(Date.now()),
})

const DURATION_LINKS = [
    { label: '7 days', days: 7 },
    { label: '14 days', days: 14 },
    { label: '30 days', days: 30 },
    { label: '60 days', days: 60 },
    { label: '1 year', days: 366 }, // 366 not 365 to account for leap year
]

const tagsFromString = (tagString: string): string[] =>
    tagString
        .split(',')
        .map(item => item.trim())
        .filter(tag => tag !== '')

const getTagsFromFormData = (formData: FormData): string[] => [
    `customer:${formData.customer}`,
    `plan:${formData.plan}`,
    ...tagsFromString(formData.tags),
]

/**
 * Displays a form to generate a new product license for a product subscription.
 *
 * For use on Sourcegraph.com by Sourcegraph teammates only.
 */
export const SiteAdminGenerateProductLicenseForSubscriptionForm: React.FunctionComponent<
    React.PropsWithChildren<Props>
> = ({ latestLicense, subscriptionID, subscriptionAccount, onGenerate, onCancel }) => {
    const labelId = 'generateLicense'

    const [formData, setFormData] = useState<FormData>(getEmptyFormData(subscriptionAccount, latestLicense))

    const onPlanChange = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(
        event => setFormData(formData => ({ ...formData, plan: event.target.value })),
        []
    )

    const useOnChange = (key: string): React.ChangeEventHandler<HTMLInputElement> =>
        useCallback<React.ChangeEventHandler<HTMLInputElement>>(
            event => setFormData(formData => ({ ...formData, [key]: event.target.value })),
            [key]
        )
    const onCustomerChange = useOnChange('customer')
    const onSFSubscriptionIDChange = useOnChange('salesforceSubscriptionID')
    const onSFOpportunityIDChange = useOnChange('salesforceOpportunityID')

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
            expiresAt: addDaysAndRoundToEndOfDay(validDays),
        }))
    }, [])
    const onExpiresAtChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event =>
            setFormData(formData => ({
                ...formData,
                expiresAt: endOfDay(
                    event.target.valueAsDate || getEmptyFormData(subscriptionAccount, latestLicense).expiresAt
                ),
            })),
        [subscriptionAccount, latestLicense]
    )

    const [generateLicense, { loading, error }] = useMutation<
        GenerateProductLicenseForSubscriptionResult['dotcom']['generateProductLicenseForSubscription'],
        GenerateProductLicenseForSubscriptionVariables
    >(GENERATE_PRODUCT_LICENSE, {
        variables: {
            productSubscriptionID: subscriptionID,
            license: {
                tags: getTagsFromFormData(formData),
                userCount: formData.userCount,
                expiresAt: Math.floor(formData.expiresAt.getTime() / 1000),
                salesforceSubscriptionID:
                    formData.salesforceSubscriptionID.trim().length > 0
                        ? formData.salesforceSubscriptionID.trim()
                        : undefined,
                salesforceOpportunityID:
                    formData.salesforceOpportunityID.trim().length > 0
                        ? formData.salesforceOpportunityID.trim()
                        : undefined,
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
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            generateLicense()
        },
        [generateLicense]
    )

    const tags = useDebounce<string[]>(tagsFromString(formData.tags), 300)

    const knownNonPlanTags = window.context.licenseInfo?.knownLicenseTags?.filter(tag => !tag.startsWith('plan:')) || []

    return (
        <Modal position="center" onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Generate new product license</H3>
            <Alert variant="info">
                Please read the{' '}
                <Link to="https://handbook.sourcegraph.com/ce/license_keys#how-to-create-a-license-key-for-a-new-prospect-or-new-customer">
                    guide for how to create a license key for a new prospect or new customer.
                </Link>
            </Alert>
            {error && <ErrorAlert error={error} />}

            <div className="site-admin-generate-product-license-for-subscription-form">
                <Form onSubmit={onSubmit}>
                    <Input
                        id="site-admin-create-product-subscription-page__customer_input"
                        label="Customer"
                        description="Name of the customer. Defaults to the Account name from the subscription."
                        type="text"
                        disabled={loading}
                        value={formData.customer || ''}
                        onChange={onCustomerChange}
                    />

                    <Input
                        id="site-admin-create-product-subscription-page__salesforce_sub_id_input"
                        label="Salesforce Subscription ID"
                        description="Subscription ID from Salesforce."
                        type="text"
                        disabled={loading}
                        value={formData.salesforceSubscriptionID}
                        onChange={onSFSubscriptionIDChange}
                    />

                    <Input
                        id="site-admin-create-product-subscription-page__salesforce_op_id_input"
                        label="Salesforce Opportunity ID"
                        description="Opportunity ID from Salesforce."
                        type="text"
                        disabled={loading}
                        value={formData.salesforceOpportunityID}
                        onChange={onSFOpportunityIDChange}
                    />

                    <Select
                        id="site-admin-create-product-subscription-page__plan_select"
                        label="Plan"
                        disabled={loading}
                        value={formData.plan}
                        onChange={onPlanChange}
                        description="Subscription plan. Required parameter."
                        className="mb-2"
                    >
                        {plans.map(plan => (
                            <option key={plan} value={plan}>
                                {plan}
                            </option>
                        ))}
                    </Select>
                    <Input
                        type="text"
                        label="Tags"
                        id="site-admin-create-product-subscription-page__tags"
                        disabled={loading}
                        value={formData.tags}
                        onChange={onTagsChange}
                        list="known-tags"
                        description="Comma separated list of tags. Tags restrict a license."
                        message={
                            <>
                                {hasUnknownTags(tags) && <UnknownTagWarning className="mb-2" />}
                                <Text className="mb-0">
                                    <ProductLicenseTags tags={tags} />
                                </Text>
                            </>
                        }
                    />
                    <datalist id="known-tags">
                        {knownNonPlanTags.map(tag => (
                            <option key={tag} value={tag}>
                                {tag}
                            </option>
                        ))}
                    </datalist>
                    <Input
                        type="number"
                        label="Users"
                        min={1}
                        id="site-admin-create-product-subscription-page__userCount"
                        disabled={loading}
                        value={formData.userCount || ''}
                        onChange={onUserCountChange}
                        description="The maximum number of users permitted on this license."
                    />
                    <Input
                        type="date"
                        label="Expires At"
                        id="site-admin-create-product-subscription-page__expiresAt"
                        min={formatDateForInput(addDaysAndRoundToEndOfDay(1))}
                        max={formatDateForInput(addDaysAndRoundToEndOfDay(2000))}
                        value={formatDateForInput(formData.expiresAt)}
                        onChange={onExpiresAtChange}
                        message={
                            <>
                                {formData.expiresAt !== null && (
                                    <ExpirationDate
                                        date={formData.expiresAt}
                                        showTime={true}
                                        showRelative={true}
                                        showPrefix={true}
                                    />
                                )}
                                <Text>
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
                                </Text>
                            </>
                        }
                    />
                    <div className="d-flex justify-content-end">
                        <Button
                            disabled={loading}
                            className="mr-2"
                            onClick={onCancel}
                            outline={true}
                            variant="secondary"
                        >
                            Cancel
                        </Button>
                        <LoaderButton
                            type="submit"
                            disabled={loading}
                            variant="primary"
                            loading={loading}
                            alwaysShowLabel={true}
                            label="Generate license"
                        />
                    </div>
                </Form>
            </div>
        </Modal>
    )
}

/**
 * Adds 1 day to the current date, then rounds it up to midnight in the client's timezone. This is a
 * generous interpretation of "valid for N days" to avoid confusion over timezones or "will it
 * expire at the beginning of the day or at the end of the day?"
 */
const addDaysAndRoundToEndOfDay = (amount: number): Date => endOfDay(addDays(Date.now(), amount))

const formatDateForInput = (date: Date): string => date.toISOString().slice(0, 10)
