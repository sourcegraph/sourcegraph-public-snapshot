import classnames from 'classnames'
import React, { forwardRef, InputHTMLAttributes, Ref } from 'react'

import { FormInput } from '../../../../../form/form-input/FormInput'
import { FormRadioInput } from '../../../../../form/form-radio-input/FormRadioInput'

import styles from './DrillDownFilters.module.scss'

interface DrillDownFiltersProps {

}

export const DrillDownFilters: React.FunctionComponent<DrillDownFiltersProps> = props => {
    const { } = props

    return (
        // eslint-disable-next-line react/forbid-elements
        <form className={styles.drilldownFilters} action="">

            <header>
                <h2>Filters by Repositories</h2>
                <button type="button" className='button btn-link'>Clear filters</button>
            </header>

            <FormRadioInput checked={true} title="Regular expression" />

            <FormInput
                as={DrillDownRegExpInput}
                prefix="repo:"
                title='Include repositories'
                placeholder='regexp-pattern' />

            <FormInput
                as={DrillDownRegExpInput}
                prefix="-repo:"
                title='Exclude repositories'
                placeholder='regexp-pattern' />
        </form>
    )
}

interface DrillDownRegExpInputProps extends InputHTMLAttributes<HTMLInputElement> {
    prefix: string
}

export const DrillDownRegExpInput = forwardRef((props: DrillDownRegExpInputProps, reference: Ref<HTMLInputElement>) => {
    const { prefix, ...inputProps } = props

    return (
        <span className={styles.regexpField}>
            <span className={styles.regexpFieldPrefix}>{prefix}</span>
            <input
                {...inputProps}
                className={classnames(inputProps.className, styles.regexpFieldInput)}
                ref={reference} />
        </span>
    )
})

DrillDownRegExpInput.displayName = 'DrillDownRegExpInput'
