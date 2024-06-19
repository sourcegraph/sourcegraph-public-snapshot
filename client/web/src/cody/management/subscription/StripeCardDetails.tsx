import React, { useCallback } from 'react'

import { CardNumberElement, CardExpiryElement, CardCvcElement } from '@stripe/react-stripe-js'
import type { StripeCardElementOptions } from '@stripe/stripe-js'

import { useTheme, Theme } from '@sourcegraph/shared/src/theme'
import { Label, Text, Grid } from '@sourcegraph/wildcard'

const useCardElementOptions = (): ((type: 'number' | 'expiry' | 'cvc') => StripeCardElementOptions) => {
    const { theme } = useTheme()

    return useCallback(
        (type: 'number' | 'expiry' | 'cvc') => ({
            ...(type === 'number' ? { disableLink: true } : {}),

            classes: {
                base: 'form-control py-2',
                focus: 'focus-visible',
                invalid: 'is-invalid',
            },

            style: {
                base: {
                    color: theme === Theme.Light ? '#262b38' : '#dbe2f0',
                },
            },
        }),
        [theme]
    )
}

interface StripeCardDetailsProps {
    onFocus?: () => void
    className?: string
}

export const StripeCardDetails: React.FC<StripeCardDetailsProps> = ({ onFocus, className }) => {
    const getOptions = useCardElementOptions()

    return (
        <div className={className}>
            <div>
                <Label className="d-block font-medium text-sm">
                    <Text className="mb-1">Card number</Text>
                    <CardNumberElement options={getOptions('number')} onFocus={onFocus} />
                </Label>
            </div>

            <Grid columnCount={2} className="mt-3 mb-0 pb-3 font-medium text-sm">
                <Label className="d-block">
                    <Text className="mb-1">Expiry date</Text>
                    <CardExpiryElement options={getOptions('expiry')} onFocus={onFocus} />
                </Label>

                <Label className="d-block font-medium text-sm">
                    <Text className="mb-1">CVC</Text>
                    <CardCvcElement options={getOptions('cvc')} onFocus={onFocus} />
                </Label>
            </Grid>
        </div>
    )
}
