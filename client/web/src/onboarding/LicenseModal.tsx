import { type FC, useCallback, useState, useEffect, type FormEvent } from 'react'

// eslint-disable-next-line id-length
import cx from 'classnames'
import { format, formatDistanceToNow } from 'date-fns'
import * as jsonc from 'jsonc-parser'
import { useNavigate } from 'react-router-dom'
import { useDebounce } from 'use-debounce'

import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Alert, Button, Form, H2, H3, Input, Label, LoadingSpinner, Modal, Text } from '@sourcegraph/wildcard'

import { BrandLogo } from '../components/branding/BrandLogo'

import { DEBOUNCE_DELAY_MS, DEFAULT_FORMAT_OPTIONS, DEFAULT_LICENSE_KEY_INFO } from './constants'
import { useUpdateLicenseKeyMutation } from './graphql-hooks'
import type { LicenseInfo, LicenseKeyModalProps, LicenseKeyInfo, LicenseKeyProps } from './types'

import styles from './LicenseModal.module.scss'

export const LicenseKeyModal: FC<LicenseKeyModalProps> = ({
    licenseKey,
    onHandleLicenseCheck,
    config,
    id,
    refetch,
}): JSX.Element => {
    const navigate = useNavigate()

    const [isValid, setIsValid] = useState(false)
    const [licenseKeyInput, setLicenseKeyInput] = useState(licenseKey.key || '')
    const [debouncedValue] = useDebounce(licenseKeyInput, DEBOUNCE_DELAY_MS)

    const onDismiss = useCallback(() => {
        onHandleLicenseCheck(true)
    }, [onHandleLicenseCheck])

    const [updateLicenseKey, { error, loading }] = useUpdateLicenseKeyMutation()

    useEffect(() => {
        const setLicenseKey = async (): Promise<void> => {
            const editFns = [
                (contents: string) => jsonc.modify(contents, ['licenseKey'], debouncedValue, DEFAULT_FORMAT_OPTIONS),
            ]
            let input = config
            for (const editFunc of editFns) {
                input = jsonc.applyEdits(config, editFunc(config))
            }
            await updateLicenseKey({ variables: { lastID: id, input } })
            setIsValid(true)
            await refetch()
        }
        if (debouncedValue.length > 10) {
            setLicenseKey().catch(() => setIsValid(false))
        }

        // When we execute the updateLicenseKey mutation, it triggers a refetch of the site configuration query.
        // As a result, we want to ensure that the useEffect doesn't run again due to the updated configuration.
        // By omitting both 'config' and 'id' from the useEffect's dependency array, we prevent unnecessary
        // re-runs of useEffect when these values change.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [updateLicenseKey, debouncedValue, refetch])

    const onSubmit = useCallback(
        async (event?: FormEvent<HTMLFormElement>) => {
            const submit = async (): Promise<void> => {
                event?.preventDefault()
                onHandleLicenseCheck(true)
                navigate('site-admin/configuration')
                await refetch()
            }
            await submit()
        },
        [navigate, refetch, onHandleLicenseCheck]
    )

    return (
        <Modal
            className={styles.modal}
            containerClassName={styles.modalContainer}
            onDismiss={() => onHandleLicenseCheck(true)}
            aria-labelledby="license-key"
        >
            <H3 className="m-0 pb-4">Upgrade your license</H3>
            <Text className="m-0 pb-3">Enter your license key to start your enterprise set up:</Text>
            {error && <Alert variant="danger">License key not recognized. Please try again.</Alert>}
            {/* eslint-disable-next-line @typescript-eslint/no-misused-promises */}
            <Form onSubmit={onSubmit}>
                <Label htmlFor="license-key">License key</Label>
                <Input
                    type="text"
                    name="license-key"
                    value={licenseKeyInput}
                    className="pb-4"
                    onChange={({ target: { value } }) => setLicenseKeyInput(value)}
                />
                <LicenseKey isValid={isValid} licenseInfo={licenseKey} />
                <div className="d-flex justify-content-end">
                    <Button className="mr-2" onClick={onDismiss} outline={true} variant="secondary" disabled={isValid}>
                        Skip for now
                    </Button>
                    <Button className={styles.submit} type="submit" disabled={!isValid} variant="primary">
                        Upgrade and start set up
                        {loading && (
                            <div data-testid="action-item-spinner">
                                <LoadingSpinner inline={false} />
                            </div>
                        )}
                    </Button>
                </div>
            </Form>
        </Modal>
    )
}

const LicenseKey: FC<LicenseKeyProps> = ({ isValid, licenseInfo }): JSX.Element => {
    const isLightTheme = useIsLightTheme()
    const { title, type, description, logo } = getLicenseKeyInfo(isValid, licenseInfo, isLightTheme)

    return (
        <div className={cx(styles.keyWrapper, 'p-4 mb-4')}>
            <div className={styles.keyContainer}>
                <div className={styles.logoContainer}>{logo}</div>
                <div>
                    <Text className="mb-1">{title}</Text>
                    <H2 className="mb-1">{type}</H2>
                    <Text className="mb-0">{description}</Text>
                </div>
            </div>
        </div>
    )
}

function getLicenseKeyInfo(isValid: boolean, licenseInfo: LicenseInfo, isLightTheme: boolean): LicenseKeyInfo {
    if (!isValid) {
        return DEFAULT_LICENSE_KEY_INFO
    }

    const { tags = [], userCount = 10, expiresAt = '' } = licenseInfo

    const expiration = new Date(expiresAt)
    const validDate = format(expiration, 'y-LL-dd')
    const remaining = formatDistanceToNow(expiration, { addSuffix: false })
    const title = tags.includes('dev') ? 'Sourcegraph Enterprise (dev-only)' : 'Sourcegraph Enterprise'
    const type = tags.includes('dev') ? 'Dev-only' : 'Enterprise'
    const numUsers = userCount > 0 ? userCount : 1
    const logo = <BrandLogo isLightTheme={isLightTheme} variant="symbol" />

    const description = `${numUsers}-user license, valid until ${validDate} (${remaining} remaining)`
    return { title, type, description, logo }
}
