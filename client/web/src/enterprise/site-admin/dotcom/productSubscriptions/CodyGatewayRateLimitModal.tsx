import React, { useCallback, useState } from 'react'

import { logger } from '@sourcegraph/common'
import { useMutation } from '@sourcegraph/http-client'
import { Button, Modal, Input, H3, Text, ErrorAlert, Form } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'
import type {
    CodyGatewayRateLimitFields,
    Scalars,
    UpdateCodyGatewayConfigResult,
    UpdateCodyGatewayConfigVariables,
} from '../../../../graphql-operations'

import { UPDATE_CODY_GATEWAY_CONFIG } from './backend'
import { ModelBadges } from './ModelBadges'
import { numberFormatter, prettyInterval } from './utils'

export interface CodyGatewayRateLimitModalProps {
    onCancel: () => void
    afterSave: () => void
    productSubscriptionID: Scalars['ID']
    current: CodyGatewayRateLimitFields | null
    mode: 'chat' | 'code' | 'embeddings'
}

export const CodyGatewayRateLimitModal: React.FunctionComponent<
    React.PropsWithChildren<CodyGatewayRateLimitModalProps>
> = ({ onCancel, afterSave, productSubscriptionID, current, mode }) => {
    const labelId = 'codyGatewayRateLimit'

    const [limit, setLimit] = useState<number>(Number(current?.limit) ?? 100)
    const onChangeLimit = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setLimit(parseInt(event.target.value, 10))
    }, [])

    const [limitInterval, setLimitInterval] = useState<number>(current?.intervalSeconds ?? 60 * 60)
    const onChangeLimitInterval = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setLimitInterval(parseInt(event.target.value, 10))
    }, [])

    const [allowedModels, setAllowedModels] = useState<string>(current?.allowedModels?.join(',') ?? '')
    const onChangeAllowedModels = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setAllowedModels(event.target.value)
    }, [])

    const [updateCodyGatewayConfig, { loading, error }] = useMutation<
        UpdateCodyGatewayConfigResult,
        UpdateCodyGatewayConfigVariables
    >(UPDATE_CODY_GATEWAY_CONFIG)

    const onSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()

            try {
                await updateCodyGatewayConfig({
                    variables: {
                        productSubscriptionID,
                        access: {
                            ...(mode === 'chat'
                                ? {
                                      chatCompletionsRateLimit: String(limit),
                                      chatCompletionsRateLimitIntervalSeconds: limitInterval,
                                      chatCompletionsAllowedModels: splitModels(allowedModels),
                                  }
                                : {}),

                            ...(mode === 'code'
                                ? {
                                      codeCompletionsRateLimit: String(limit),
                                      codeCompletionsRateLimitIntervalSeconds: limitInterval,
                                      codeCompletionsAllowedModels: splitModels(allowedModels),
                                  }
                                : {}),

                            ...(mode === 'embeddings'
                                ? {
                                      embeddingsRateLimit: String(limit),
                                      embeddingsRateLimitIntervalSeconds: limitInterval,
                                      embeddingsAllowedModels: splitModels(allowedModels),
                                  }
                                : {}),
                        },
                    },
                })

                afterSave()
            } catch (error) {
                // Non-request error. API errors will be available under `error` above.
                logger.error(error)
            }
        },
        [updateCodyGatewayConfig, productSubscriptionID, limit, limitInterval, afterSave, allowedModels, mode]
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
                        value={limitInterval}
                        onChange={onChangeLimitInterval}
                        message={
                            <>
                                {numberFormatter.format(BigInt(limit))} {mode === 'embeddings' ? 'tokens' : 'requests'}{' '}
                                per {prettyInterval(limitInterval)}
                            </>
                        }
                    />
                </div>
                <div className="form-group">
                    <Input
                        id="allowedModels"
                        name="allowedModels"
                        type="text"
                        autoComplete="off"
                        spellCheck="false"
                        required={true}
                        disabled={loading}
                        min={1}
                        label="Allowed models"
                        description="Comma separated list of the models the subscription can use. This normally doesn't need to be changed."
                        value={allowedModels}
                        onChange={onChangeAllowedModels}
                        message={
                            <ModelBadges
                                models={splitModels(allowedModels)}
                                mode={mode === 'embeddings' ? 'embeddings' : 'completions'}
                            />
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

function splitModels(allowedModels: string): string[] {
    if (allowedModels === '') {
        return []
    }
    return allowedModels.split(',').map(model => model.trim())
}
