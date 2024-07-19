import { type ComponentProps, type FunctionComponent, useCallback, useMemo, useState } from 'react'

import { Input } from '@sourcegraph/wildcard'

/**
 * An <input> component that constrains the value to match a pattern. This is useful for inputs for
 * the names of things that need to be clean path components, containing no spaces and limited
 * punctuation.
 */
export const PatternConstrainedInput: FunctionComponent<
    {
        value: string
        pattern: string
        replaceSpaces?: boolean
        onChange: (value: string, isValid: boolean) => void
    } & Pick<
        ComponentProps<typeof Input>,
        'label' | 'name' | 'required' | 'disabled' | 'autoComplete' | 'autoCapitalize'
    >
> = ({ value, pattern, replaceSpaces, onChange: parentOnChange, ...props }) => {
    // HTMLInputElement.pattern uses the 'v' flag:
    // https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes/pattern.
    const patternRegExp = useMemo(() => new RegExp(`^${pattern}$`, 'v'), [pattern])

    const [isValid, setIsValid] = useState<boolean>()
    const onChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => {
            const newValue = replaceSpaces ? event.target.value.replaceAll(' ', '-') : event.target.value
            const isValid = patternRegExp.test(newValue)
            parentOnChange(newValue, isValid)
            setIsValid(isValid)
        },
        [parentOnChange, patternRegExp, replaceSpaces]
    )

    return (
        <Input
            value={value}
            onChange={onChange}
            pattern={pattern}
            status={isValid === undefined ? undefined : isValid ? 'valid' : 'error'}
            {...props}
        />
    )
}
