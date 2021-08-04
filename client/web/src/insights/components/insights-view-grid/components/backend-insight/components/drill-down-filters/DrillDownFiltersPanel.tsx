import classnames from 'classnames'
import React, { forwardRef, InputHTMLAttributes, PropsWithChildren, Ref } from 'react'

import { TruncatedText } from '../../../../../../pages/dashboards/dashboard-page/components/dashboard-select/components/trancated-text/TrancatedText'
import { FormInput } from '../../../../../form/form-input/FormInput'
import { FormRadioInput } from '../../../../../form/form-radio-input/FormRadioInput'
import { useForm } from '../../../../../form/hooks/useForm'

import styles from './DrillDownFiltersPanel.module.scss'

interface DrillDownFilters {
    includeRepoRegExp: string
    excludeRepoRegExp: string
}

const DEFAULT_FILTERS: DrillDownFilters = {
    excludeRepoRegExp: '',
    includeRepoRegExp: '',
}

interface DrillDownFiltersPanelProps {
    className?: string
    filters?: DrillDownFilters
}

export const DrillDownFiltersPanel: React.FunctionComponent<DrillDownFiltersPanelProps> = props => {
    const { className, filters = DEFAULT_FILTERS } = props

    const { ref } = useForm<DrillDownFilters>({
        initialValues: filters,
        onChange: values => console.log(values),
    })

    return (
        // eslint-disable-next-line react/forbid-elements
        <form ref={ref} className={classnames(className, styles.drilldownFilters)}>
            <header className="d-flex align-items-center">
                <h4 className="m-0">Filters by Repositories</h4>
                <button type="button" className="btn btn-link ml-auto">
                    Clear filters
                </button>
            </header>

            <hr className="ml-n3 mr-n3" />

            <FormRadioInput checked={true} title="Regular expression" className="pt-3 pb-3" />

            <FormInput
                as={DrillDownRegExpInput}
                autoFocus={true}
                autofocus={true}
                prefix="repo:"
                title={<LabelWithReset>Include repositories</LabelWithReset>}
                placeholder="regexp-pattern"
                className="mb-4"
            />

            <FormInput
                as={DrillDownRegExpInput}
                prefix="-repo:"
                title={<LabelWithReset>Exclude repositories</LabelWithReset>}
                placeholder="regexp-pattern"
            />
        </form>
    )
}

interface LabelWithResetProps {
    onReset?: () => void
}

const LabelWithReset: React.FunctionComponent<PropsWithChildren<LabelWithResetProps>> = props => (
    <span className="d-flex align-items-center">
        <TruncatedText>{props.children}</TruncatedText>
        <button
            type="button"
            className="btn btn-link ml-auto pt-0 pb-0 pr-0 font-weight-normal"
            onClick={props.onReset}
        >
            Reset
        </button>
    </span>
)

interface DrillDownRegExpInputProps extends InputHTMLAttributes<HTMLInputElement> {
    prefix: string
}

export const DrillDownRegExpInput = forwardRef((props: DrillDownRegExpInputProps, reference: Ref<HTMLInputElement>) => {
    const { prefix, ...inputProps } = props

    return (
        <span className={classnames(styles.regexpField, 'w-100')}>
            <span className={styles.regexpFieldPrefix}>{prefix}</span>
            <input
                {...inputProps}
                className={classnames(inputProps.className, styles.regexpFieldInput)}
                ref={reference}
            />
        </span>
    )
})

DrillDownRegExpInput.displayName = 'DrillDownRegExpInput'
