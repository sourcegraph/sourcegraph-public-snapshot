import React, { forwardRef, InputHTMLAttributes, Ref } from 'react'

import { FormRadioInput } from '../../../../../form/form-radio-input/FormRadioInput';
import { FormInput } from '../../../../../form/form-input/FormInput';

interface DrillDownFiltersProps {

}

export const DrillDownFilters: React.FunctionComponent<DrillDownFiltersProps> = props => {
    const {} = props;

    return (
        // eslint-disable-next-line react/forbid-elements
        <form action="">

            <header>
                <h2>Filters by Repositories</h2>
                <button type="button" className='button btn-link'>Clear filters</button>
            </header>

            <FormRadioInput checked={true} title="Regular expression" />

            <FormInput
                as={DrillDownRegExpInput}
                prefix="-repo:"/>
        </form>
    )
}

interface DrillDownRegExpInputProps extends InputHTMLAttributes<HTMLInputElement> {
    prefix: string
}

export const DrillDownRegExpInput = forwardRef((props: DrillDownRegExpInputProps, reference: Ref<HTMLInputElement>) => {
    const { prefix, ...inputProps } = props;

    return (
        <span>
            <span>{ prefix }</span>
            <input {...inputProps} ref={reference}/>
        </span>
  )
})

DrillDownRegExpInput.displayName = 'DrillDownRegExpInput'
