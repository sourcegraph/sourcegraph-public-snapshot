import { useCallback, useState } from 'react'

import { mdiPencil, mdiTrashCan } from '@mdi/js'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { logger } from '@sourcegraph/common'
import { useMutation } from '@sourcegraph/http-client'
import { LLMProxyRateLimitSource } from '@sourcegraph/shared/src/graphql-operations'
import {
    H3,
    ProductStatusBadge,
    Container,
    Text,
    H4,
    ErrorAlert,
    LoadingSpinner,
    Button,
    Icon,
    Badge,
    Tooltip,
    Label,
} from '@sourcegraph/wildcard'

import { CopyableText } from '../../../../components/CopyableText'
import {
    LLMProxyAccessFields,
    Scalars,
    UpdateLLMProxyConfigResult,
    UpdateLLMProxyConfigVariables,
} from '../../../../graphql-operations'

import { UPDATE_LLM_PROXY_CONFIG } from './backend'
import { LLMProxyRateLimitModal } from './LlmProxyRateLimitModal'
import { prettyInterval } from './utils'

interface Props {
    productSubscriptionID: Scalars['ID']
    currentSourcegraphAccessToken: string | null
    accessTokenError?: Error
    viewerCanAdminister: boolean
    refetchSubscription: () => void
    llmProxyAccess: LLMProxyAccessFields
}

export const CodyServicesSection: React.FunctionComponent<Props> = ({
    productSubscriptionID,
    viewerCanAdminister,
    currentSourcegraphAccessToken,
    accessTokenError,
    refetchSubscription,
    llmProxyAccess,
}) => {
    const [showRateLimitConfigModal, setShowRateLimitConfigModal] = useState<boolean>(false)

    const [updateLLMProxyConfig, { loading: updateLLMProxyConfigLoading, error: updateLLMProxyConfigError }] =
        useMutation<UpdateLLMProxyConfigResult, UpdateLLMProxyConfigVariables>(UPDATE_LLM_PROXY_CONFIG)

    const onRemoveRateLimitOverride = useCallback(async () => {
        try {
            await updateLLMProxyConfig({
                variables: {
                    productSubscriptionID,
                    llmProxyAccess: { rateLimit: 0, rateLimitIntervalSeconds: 0 },
                },
            })
            refetchSubscription()
        } catch (error) {
            logger.error(error)
        }
    }, [productSubscriptionID, refetchSubscription, updateLLMProxyConfig])

    const onToggleCompletions = useCallback(
        async (value: boolean) => {
            try {
                await updateLLMProxyConfig({
                    variables: {
                        productSubscriptionID,
                        llmProxyAccess: { enabled: value },
                    },
                })
                refetchSubscription()
            } catch (error) {
                logger.error(error)
            }
        },
        [productSubscriptionID, refetchSubscription, updateLLMProxyConfig]
    )

    const afterSaveRateLimit = useCallback(() => {
        refetchSubscription()
        setShowRateLimitConfigModal(false)
    }, [refetchSubscription])

    return (
        <>
            <H3>
                Cody services <ProductStatusBadge status="experimental" />
            </H3>
            <Container className="mb-3">
                <H4>Access token</H4>
                <Text className="mb-2">Access tokens can be used for LLM-proxy access - coming soon!</Text>
                {currentSourcegraphAccessToken && (
                    <CopyableText
                        label="Access token"
                        secret={true}
                        flex={true}
                        text={currentSourcegraphAccessToken}
                        className="mb-2"
                    />
                )}
                {accessTokenError && <ErrorAlert error={accessTokenError} className="mb-0" />}

                {currentSourcegraphAccessToken && (
                    <>
                        <H4>Completions</H4>

                        <div className="form-group mb-0">
                            {updateLLMProxyConfigError && <ErrorAlert error={updateLLMProxyConfigError} />}
                            <Label className="mb-0">
                                <Toggle
                                    id="llm-proxy-enabled"
                                    value={llmProxyAccess.enabled}
                                    disabled={updateLLMProxyConfigLoading || !viewerCanAdminister}
                                    onToggle={onToggleCompletions}
                                    className="mr-1 align-text-bottom"
                                />
                                Access to hosted completions (LLM-proxy)
                                {updateLLMProxyConfigLoading && (
                                    <>
                                        {' '}
                                        <LoadingSpinner />
                                    </>
                                )}
                            </Label>
                        </div>

                        {llmProxyAccess.enabled && (
                            <div className="form-group mt-2 mb-0">
                                <Label className="mb-2">Rate limit</Label>
                                <Text className="mb-0 d-flex align-items-baseline">
                                    {llmProxyAccess.rateLimit !== null && (
                                        <>
                                            <LLMProxyRateLimitSourceBadge
                                                source={llmProxyAccess.rateLimit.source}
                                                className="mr-2"
                                            />
                                            {llmProxyAccess.rateLimit.limit} requests /{' '}
                                            {prettyInterval(llmProxyAccess.rateLimit.intervalSeconds)}
                                            {viewerCanAdminister && (
                                                <>
                                                    <Button
                                                        size="sm"
                                                        variant="link"
                                                        aria-label="Edit rate limit"
                                                        className="ml-1"
                                                        onClick={() => setShowRateLimitConfigModal(true)}
                                                    >
                                                        <Icon aria-hidden={true} svgPath={mdiPencil} />
                                                    </Button>
                                                    {llmProxyAccess.rateLimit.source ===
                                                        LLMProxyRateLimitSource.OVERRIDE && (
                                                        <Tooltip content="Remove rate limit override">
                                                            <Button
                                                                size="sm"
                                                                variant="link"
                                                                aria-label="Remove rate limit override"
                                                                className="ml-1"
                                                                onClick={onRemoveRateLimitOverride}
                                                            >
                                                                <Icon
                                                                    aria-hidden={true}
                                                                    svgPath={mdiTrashCan}
                                                                    className="text-danger"
                                                                />
                                                            </Button>
                                                        </Tooltip>
                                                    )}
                                                </>
                                            )}
                                        </>
                                    )}
                                </Text>
                            </div>
                        )}
                    </>
                )}
            </Container>

            {showRateLimitConfigModal && (
                <LLMProxyRateLimitModal
                    productSubscriptionID={productSubscriptionID}
                    afterSave={afterSaveRateLimit}
                    current={llmProxyAccess.rateLimit}
                    onCancel={() => setShowRateLimitConfigModal(false)}
                />
            )}
        </>
    )
}

export const LLMProxyRateLimitSourceBadge: React.FunctionComponent<{
    source: LLMProxyRateLimitSource
    className?: string
}> = ({ source, className }) => {
    switch (source) {
        case LLMProxyRateLimitSource.OVERRIDE:
            return (
                <Tooltip content="The limit has been specified by a custom override">
                    <Badge variant="primary" className={className}>
                        Override
                    </Badge>
                </Tooltip>
            )
        case LLMProxyRateLimitSource.PLAN:
            return (
                <Tooltip content="The limit is derived from the current subscription plan">
                    <Badge variant="primary" className={className}>
                        Plan
                    </Badge>
                </Tooltip>
            )
    }
}
