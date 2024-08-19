import React, { useCallback, useState } from 'react'

import { Duration } from '@bufbuild/protobuf'

import { logger } from '@sourcegraph/common'
import { Button, Modal, Input, H3, Text, ErrorAlert, Form, Alert } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'

import { type EnterprisePortalEnvironment, useUpdateCodyGatewayAccess } from './enterpriseportal'
import type { CodyGatewayRateLimit } from './enterpriseportalgen/codyaccess_pb'
import { numberFormatter, prettyInterval } from './utils'

export interface CodyGatewayRateLimitModalProps {
    enterprisePortalEnvironment: EnterprisePortalEnvironment
    onCancel: () => void
    afterSave: () => void
    productSubscriptionUUID: string
    current: CodyGatewayRateLimit | undefined
    mode: 'chat' | 'code' | 'embeddings'
}

export const CodyGatewayRateLimitModal: React.FunctionComponent<
    React.PropsWithChildren<CodyGatewayRateLimitModalProps>
> = ({ onCancel, afterSave, productSubscriptionUUID, current, mode, enterprisePortalEnvironment }) => {
    const labelId = 'codyGatewayRateLimit'

    const [limit, setLimit] = useState<bigint>(current?.limit ?? BigInt(100))
    const onChangeLimit = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setLimit(BigInt(parseInt(event.target.value || '0', 10)))
    }, [])

    /**
     * Interval is tracked in seconds in state.
     */
    const [limitInterval, setLimitInterval] = useState<bigint>(current?.intervalDuration?.seconds ?? BigInt(60 * 60))
    const onChangeLimitInterval = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setLimitInterval(BigInt(parseInt(event.target.value || '0', 10)))
    }, [])

    const {
        error,
        isPending: loading,
        mutateAsync: updateCodyGatewayAccess,
    } = useUpdateCodyGatewayAccess(enterprisePortalEnvironment)

    const onSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()

            try {
                // Explicitly indicate what should be updated - optional, but
                // just to be clear/safe.
                const paths: string[] = []
                switch (mode) {
                    case 'chat': {
                        paths.push('chat_completions_rate_limit.limit', 'chat_completions_rate_limit.interval_duration')
                        break
                    }
                    case 'code': {
                        paths.push('code_completions_rate_limit.limit', 'code_completions_rate_limit.interval_duration')
                        break
                    }
                    case 'embeddings': {
                        paths.push('embeddings_rate_limit.limit', 'embeddings_rate_limit.interval_duration')
                        break
                    }
                }
                const rateLimit = {
                    limit,
                    intervalDuration: new Duration({ seconds: limitInterval }),
                }
                await updateCodyGatewayAccess({
                    updateMask: { paths },
                    access: {
                        subscriptionId: productSubscriptionUUID,
                        /**
                         * All non-zero fields are included in the update
                         */
                        ...(mode === 'chat'
                            ? {
                                  chatCompletionsRateLimit: rateLimit,
                              }
                            : {}),

                        ...(mode === 'code'
                            ? {
                                  codeCompletionsRateLimit: rateLimit,
                              }
                            : {}),

                        ...(mode === 'embeddings'
                            ? {
                                  embeddingsRateLimit: rateLimit,
                              }
                            : {}),
                    },
                })

                afterSave()
            } catch (error) {
                // Non-request error. API errors will be available under `error` above.
                logger.error(error)
            }
        },
        [updateCodyGatewayAccess, productSubscriptionUUID, limit, limitInterval, afterSave, mode]
    )

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>
                Configure custom{' '}
                {mode === 'chat'
                    ? 'chat request'
                    : mode === 'code'
                    ? 'code completion request'
                    : 'embeddings generation'}{' '}
                rate limit override for Cody Gateway access
            </H3>
            <Text>
                Overrides take precedence over the default rate limits, which are based on the active licence's plan and
                user count.
            </Text>
            <Alert variant="warning">
                Rate limit overrides are static: for example, they must be updated manually, or removed, to accomodate
                an increase in user count.
            </Alert>

            {error && <ErrorAlert error={error} />}

            <Form onSubmit={onSubmit}>
                <div className="form-group">
                    <Input
                        id="limit"
                        name="limit"
                        autoComplete="off"
                        inputClassName="mb-2"
                        className="mb-0"
                        required={true}
                        disabled={loading}
                        spellCheck="false"
                        type="number"
                        min={1}
                        value={Number(limit)}
                        onChange={onChangeLimit}
                        label={mode === 'embeddings' ? 'Number of tokens embedded' : 'Number of requests'}
                    />
                </div>
                <div className="form-group">
                    <Input
                        id="limitInterval"
                        name="limitInterval"
                        type="number"
                        autoComplete="off"
                        spellCheck="false"
                        required={true}
                        disabled={loading}
                        min={1}
                        label="Rate limit interval"
                        description="The interval is defined in seconds. See below for a pretty-printed version."
                        value={Number(limitInterval)}
                        onChange={onChangeLimitInterval}
                        message={
                            <>
                                {numberFormatter.format(BigInt(limit || 0))}{' '}
                                {mode === 'embeddings' ? 'tokens' : 'requests'} per{' '}
                                {prettyInterval(Number(limitInterval))}
                            </>
                        }
                    />
                </div>
                <div className="d-flex justify-content-end">
                    <Button disabled={loading} className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                        Cancel
                    </Button>
                    <LoaderButton
                        type="submit"
                        disabled={loading || limit <= 0 || limitInterval <= 0}
                        variant="primary"
                        loading={loading}
                        alwaysShowLabel={true}
                        label="Save"
                    />
                </div>
            </Form>
        </Modal>
    )
}
