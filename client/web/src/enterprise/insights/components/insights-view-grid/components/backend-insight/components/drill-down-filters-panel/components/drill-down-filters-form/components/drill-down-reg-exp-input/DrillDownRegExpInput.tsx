import classNames from 'classnames'
import React, { forwardRef, InputHTMLAttributes, PropsWithChildren, Ref } from 'react'

import { Button, FlexTextArea } from '@sourcegraph/wildcard'

import { TruncatedText } from '../../../../../../../../../../pages/dashboards/dashboard-page/components/dashboard-select/components/trancated-text/TrancatedText'

import styles from './DrillDownRegExpInput.module.scss'

interface DrillDownRegExpInputProps extends InputHTMLAttributes<HTMLInputElement> {
    prefix: string
}

export const DrillDownRegExpInput = forwardRef((props: DrillDownRegExpInputProps, reference: Ref<HTMLInputElement>) => {
    const { prefix, ...inputProps } = props

    return (
        <span className="d-flex w-100">
            <span className={classNames(styles.prefixText, 'text-monospace')}>{prefix}</span>
            <FlexTextArea
                {...inputProps}
                containerClassName="w-100"
                className={classNames(inputProps.className, styles.input)}
                ref={reference}
            />
        </span>
    )
})

DrillDownRegExpInput.displayName = 'DrillDownRegExpInput'

interface LabelWithResetProps {
    onReset?: () => void
}

export const LabelWithReset: React.FunctionComponent<PropsWithChildren<LabelWithResetProps>> = props => (
    <span className="d-flex align-items-center">
        <TruncatedText>{props.children}</TruncatedText>
        <Button variant="link" className="ml-auto pt-0 pb-0 pr-0 font-weight-normal" onClick={props.onReset}>
            Reset
        </Button>
    </span>
)
