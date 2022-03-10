import { StylesConfig } from 'react-select'

export const STYLES: StylesConfig = {
    clearIndicator: provided => ({
        ...provided,
        padding: '0 0.125rem',
    }),
    control: (provided, state) => ({
        ...provided,
        // Styles here replicate the styles of `wildcard/Select`
        backgroundColor: state.isDisabled ? 'var(--input-disabled-bg)' : 'var(--input-bg)',
        borderColor: state.selectProps.isValid
            ? 'var(--success)'
            : state.selectProps.isValid === false
            ? 'var(--danger)'
            : state.isFocused
            ? state.theme.colors.primary
            : 'var(--input-border-color)',
        boxShadow: state.isFocused
            ? // These are stolen from `wildcard/Input` and `wildcard/Select`, which seem to come from Bootstrap
              state.selectProps.isValid
                ? 'var(--input-focus-box-shadow-valid)'
                : state.selectProps.isValid === false
                ? 'var(--input-focus-box-shadow-invalid)'
                : 'var(--input-focus-box-shadow)'
            : undefined,
        '&:hover': {
            borderColor: undefined,
        },
    }),
    dropdownIndicator: provided => ({
        ...provided,
        padding: '0 0.125rem',
    }),
    indicatorSeparator: (provided, state) => ({
        ...provided,
        backgroundColor: state.hasValue ? 'var(--input-border-color)' : 'transparent',
    }),
    input: provided => ({
        ...provided,
        color: 'var(--input-color)',
        margin: '0 0.125rem',
        padding: 0,
    }),
    menu: provided => ({
        ...provided,
        background: 'var(--dropdown-bg)',
        padding: '0.25rem 0',
        margin: '0.125rem 0 0',
        dropShadow: 'var(--dropdown-shadow)',
    }),
    menuList: provided => ({
        ...provided,
        padding: 0,
    }),
    multiValueRemove: (provided, state) => ({
        ...provided,
        backgroundColor: 'transparent',
        boxShadow: state.isFocused ? 'var(--input-focus-box-shadow)' : undefined,
        ':hover': {
            ...provided[':hover'],
            backgroundColor: 'transparent',
            color: undefined,
        },
    }),
    noOptionsMessage: provided => ({
        ...provided,
        color: 'var(--input-placeholder-color)',
    }),
    option: (provided, state) => ({
        ...provided,
        backgroundColor: state.isSelected
            ? state.isFocused
                ? 'var(--primary-3)'
                : state.theme.colors.primary
            : state.isFocused
            ? 'var(--dropdown-link-hover-bg)'
            : undefined,
        color: state.isSelected ? 'var(--light-text)' : undefined,
        ':hover': {
            cursor: 'pointer',
        },
        ':active': {
            backgroundColor: state.isSelected
                ? state.isFocused
                    ? 'var(--primary-3)'
                    : state.theme.colors.primary
                : state.isFocused
                ? 'var(--dropdown-link-hover-bg)'
                : undefined,
        },
    }),
    placeholder: (provided, state) => ({
        ...provided,
        color: state.isDisabled ? 'var(--gray-06)' : 'var(--input-placeholder-color)',
    }),
    valueContainer: provided => ({
        ...provided,
        padding: '0.125rem 0.125rem 0.125rem 0.75rem',
    }),
}
