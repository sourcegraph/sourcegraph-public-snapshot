import { FC } from 'react'

import { mdiCheckCircle, mdiCloseCircle } from '@mdi/js'
import classNames from 'classnames'
import { reverse, sortBy } from 'lodash'

import { pluralize } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { Alert, Text, H3, Icon, Label, Input, useDebounce, LoadingSpinner, Link, Grid } from '@sourcegraph/wildcard'

import { GetLicenseInfoResult, GetLicenseInfoVariables } from '../../../../graphql-operations'
import { GET_LICENSE_INFO } from '../../../backend'

interface LicenseKeyFormProps {
    className?: string
    licenseKey?: string
    onLicenseKeyChange: (newContents: string) => void
}
export const LicenseKeyForm: FC<LicenseKeyFormProps> = ({ className, licenseKey = '', onLicenseKeyChange }) => {
    const debouncedLicenseKey = useDebounce(licenseKey, 500)
    const { data, loading, error } = useQuery<GetLicenseInfoResult, GetLicenseInfoVariables>(GET_LICENSE_INFO, {
        variables: {
            licenseKey: debouncedLicenseKey || null,
        },
    })

    const info = data?.licenseInfo
    const features = reverse(sortBy(info?.features ?? [], feature => feature.enabled))

    return (
        <div className={className}>
            <Text>A valid license key from your Sourcegraph contact enables many Enterprise-specific features.</Text>
            {loading && <LoadingSpinner />}
            {!debouncedLicenseKey && (
                <Alert variant="warning">
                    Add your license key to get full access to your Sourcegraph instance. Don’t have one?{' '}
                    <Link to="https://about.sourcegraph.com/contact/request-info">Reach out to our team</Link>.
                </Alert>
            )}
            {error && <Alert variant="danger">{error.message}</Alert>}
            {info && (
                <div className="p-2 mb-3">
                    <H3>
                        <Text as="span" weight="regular">
                            Plan:
                        </Text>{' '}
                        {info?.plan}{' '}
                        <Text as="span" weight="regular">
                            (
                            {info.userCountRestricted
                                ? `${info.userCount} ${pluralize('user', info.userCount)}`
                                : 'Dynamically scalable'}
                            )
                        </Text>
                    </H3>
                    <Text className="mb-2">Features available:</Text>
                    <Grid columnCount={2} spacing={0.5} className="pl-2">
                        {features.map(feature => (
                            <Text className="m-0" key={feature.name}>
                                <Icon
                                    svgPath={feature.enabled ? mdiCheckCircle : mdiCloseCircle}
                                    className={classNames('mr-1', {
                                        'text-success': feature.enabled,
                                        'text-danger': !feature.enabled,
                                    })}
                                    aria-label={feature.enabled ? 'Available' : 'Unavailable'}
                                />
                                {feature.name}
                            </Text>
                        ))}
                    </Grid>
                </div>
            )}

            <Label htmlFor="license-key">License Key</Label>
            <Input
                id="license-key"
                placeholder="Paste your license key here"
                onChange={event => onLicenseKeyChange(event.target.value)}
                value={licenseKey}
                className="mb-2"
            />
        </div>
    )
}
