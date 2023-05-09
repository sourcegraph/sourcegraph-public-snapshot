import React, { useCallback, useState } from 'react'

import { logger } from '@sourcegraph/common'
import { useMutation } from '@sourcegraph/http-client'
import { Button, Modal, Input, H3, Text, ErrorAlert, Form } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'
import {
    LLMProxyRateLimitFields,
    Scalars,
    UpdateLLMProxyConfigResult,
    UpdateLLMProxyConfigVariables,
} from '../../../../graphql-operations'

import { UPDATE_LLM_PROXY_CONFIG } from './backend'
import { prettyInterval } from './utils'

export interface LLMProxyRateLimitModalProps {
    onCancel: () => void
    afterSave: () => void
    productSubscriptionID: Scalars['ID']
    current: LLMProxyRateLimitFields | null
}

export const LLMProxyRateLimitModal: React.FunctionComponent<React.PropsWithChildren<LLMProxyRateLimitModalProps>> = ({
    onCancel,
    afterSave,
    productSubscriptionID,
    current,
}) => {
    const labelId = 'llmProxyRateLimit'

    const [limit, setLimit] = useState<number>(current?.limit ?? 100)
    const onChangeLimit = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setLimit(parseInt(event.target.value, 10))
    }, [])

    const [limitInterval, setLimitInterval] = useState<number>(current?.intervalSeconds ?? 60 * 60)
    const onChangeLimitInterval = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setLimitInterval(parseInt(event.target.value, 10))
    }, [])

    const [updateLLMProxyConfig, { loading, error }] = useMutation<
        UpdateLLMProxyConfigResult,
        UpdateLLMProxyConfigVariables
    >(UPDATE_LLM_PROXY_CONFIG)

    const onSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()

            try {
                await updateLLMProxyConfig({
                    variables: {
                        productSubscriptionID,
                        llmProxyAccess: {
                            rateLimit: limit,
                            rateLimitIntervalSeconds: limitInterval,
                        },
                    },
                })

                afterSave()
            } catch (error) {
                // Non-request error. API errors will be available under `error` above.
                logger.error(error)
            }
        },
        [updateLLMProxyConfig, productSubscriptionID, limit, limitInterval, afterSave]
    )

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Configure rate limit for LLM proxy</H3>
            <Text>
                LLM proxy is a Sourcegraph managed service that allows customer instances to talk to upstream LLMs under
                our negotiated terms in a safe manner.
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
                        label="Number of requests"
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
                        value={limitInterval}
                        onChange={onChangeLimitInterval}
                        message={
                            <>
                                {limit} requests per {prettyInterval(limitInterval!)}
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
