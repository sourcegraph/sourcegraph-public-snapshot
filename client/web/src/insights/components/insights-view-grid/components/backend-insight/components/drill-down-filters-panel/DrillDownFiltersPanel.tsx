import classnames from 'classnames'
import React, { forwardRef, InputHTMLAttributes, PropsWithChildren, Ref } from 'react'

import { TruncatedText } from '../../../../../../pages/dashboards/dashboard-page/components/dashboard-select/components/trancated-text/TrancatedText'
import { FormInput } from '../../../../../form/form-input/FormInput'
import { FormRadioInput } from '../../../../../form/form-radio-input/FormRadioInput'
import { useField } from '../../../../../form/hooks/useField'
import { FormChangeEvent, useForm } from '../../../../../form/hooks/useForm'

import styles from './DrillDownFiltersPanel.module.scss'
import { DrillDownFilters, DrillDownFiltersMode, EMPTY_DRILLDOWN_FILTERS } from './types'
import { validRegexp } from './validators'

interface DrillDownFiltersPanelProps {
    className?: string
    initialFiltersValue?: DrillDownFilters
    onFiltersChange?: (filters: FormChangeEvent<DrillDownFilters>) => void
}

export const DrillDownFiltersPanel: React.FunctionComponent<DrillDownFiltersPanelProps> = props => {
    const { className, initialFiltersValue = EMPTY_DRILLDOWN_FILTERS, onFiltersChange } = props

    const { ref, formAPI } = useForm<DrillDownFilters>({
        initialValues: initialFiltersValue,
        onChange: onFiltersChange,
    })

    const mode = useField({
        name: 'mode',
        formApi: formAPI,
    })

    const includeRegex = useField({
        name: 'includeRepoRegex',
        formApi: formAPI,
        validators: { sync: validRegexp },
    })

    const excludeRegex = useField({
        name: 'excludeRepoRegex',
        formApi: formAPI,
        validators: { sync: validRegexp },
    })

    const handleFiltersReset = (): void => {
        includeRegex.input.onChange('')
        excludeRegex.input.onChange('')
    }

    return (
        // eslint-disable-next-line react/forbid-elements
        <form ref={ref} className={classnames(className, styles.drilldownFilters)}>
            <header className="d-flex align-items-center">
                <h4 className="m-0">Filters by Repositories</h4>
                <button type="button" className="btn btn-link ml-auto" onClick={handleFiltersReset}>
                    Clear filters
                </button>
            </header>

            <hr className="ml-n3 mr-n3" />

            <FormRadioInput
                title="Regular expression"
                className="pt-3 pb-3"
                name="mode"
                value={DrillDownFiltersMode.Regex}
                checked={mode.input.value === DrillDownFiltersMode.Regex}
                onChange={mode.input.onChange}
            />

            <FormInput
                as={DrillDownRegExpInput}
                autoFocus={true}
                prefix="repo:"
                title={
                    <LabelWithReset onReset={() => includeRegex.input.onChange('')}>
                        Include repositories
                    </LabelWithReset>
                }
                placeholder="regexp-pattern"
                className="mb-4"
                valid={includeRegex.meta.dirty && includeRegex.meta.validState === 'VALID'}
                error={includeRegex.meta.dirty && includeRegex.meta.error}
                {...includeRegex.input}
            />

            <FormInput
                as={DrillDownRegExpInput}
                prefix="-repo:"
                title={
                    <LabelWithReset onReset={() => excludeRegex.input.onChange('')}>
                        Exclude repositories
                    </LabelWithReset>
                }
                placeholder="regexp-pattern"
                valid={excludeRegex.meta.dirty && excludeRegex.meta.validState === 'VALID'}
                error={excludeRegex.meta.dirty && excludeRegex.meta.error}
                {...excludeRegex.input}
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
