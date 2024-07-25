import React, { useCallback, useState } from 'react'

import { Duration } from '@bufbuild/protobuf'

import { logger } from '@sourcegraph/common'
import { Button, Modal, Input, H3, Text, ErrorAlert, Form } from '@sourcegraph/wildcard'

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

    const [limit, setLimit] = useState<number>(Number(current?.limit) ?? 100)
    const onChangeLimit = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setLimit(parseInt(event.target.value || '0', 10))
    }, [])

    const [limitInterval, setLimitInterval] = useState<Duration>(
        current?.intervalDuration ?? new Duration({ seconds: BigInt(60 * 60) })
    )
    const onChangeLimitInterval = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setLimitInterval(new Duration({ seconds: BigInt(parseInt(event.target.value || '0', 10)) }))
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
                await updateCodyGatewayAccess({
                    updateMask: { paths },
                    access: {
                        subscriptionId: productSubscriptionUUID,
                        /**
                         * All non-zero fields are included in the update
                         */
                        ...(mode === 'chat'
                            ? {
                                  chatCompletionsRateLimit: {
                                      limit: BigInt(limit),
                                      intervalDuration: limitInterval,
                                  },
                              }
                            : {}),

                        ...(mode === 'code'
                            ? {
                                  codeCompletionsRateLimit: {
                                      limit: BigInt(limit),
                                      intervalDuration: limitInterval,
                                  },
                              }
                            : {}),

                        ...(mode === 'embeddings'
                            ? {
                                  embeddingsRateLimit: {
                                      limit: BigInt(limit),
                                      intervalDuration: limitInterval,
                                  },
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
                Configure{' '}
                {mode === 'chat'
                    ? 'chat request'
                    : mode === 'code'
                    ? 'code completion request'
                    : 'embeddings generation'}{' '}
                rate limit for Cody Gateway
            </H3>
            <Text>
                Cody Gateway is a Sourcegraph managed service that allows customer instances to talk to upstream LLMs
                and generate embeddings under our negotiated terms with third party providers in a safe manner.
            </Text>

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
                        value={limit}
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
                        value={Number(limitInterval.seconds)}
                        onChange={onChangeLimitInterval}
                        message={
                            <>
                                {numberFormatter.format(BigInt(limit || 0))}{' '}
                                {mode === 'embeddings' ? 'tokens' : 'requests'} per{' '}
                                {prettyInterval(Number(limitInterval.seconds))}
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
                        disabled={loading || limit <= 0 || Number(limitInterval.seconds) <= 0}
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
