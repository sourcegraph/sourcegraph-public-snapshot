import { ThemeConfig } from 'react-select'

export const THEME: ThemeConfig = theme => ({
    ...theme,
    borderRadius: 3,
    // Each identifiable instance of `Select` components using a theme color has been
    // overwritten by `STYLES` in order to switch to light mode/dark mode. These colors
    // defined here only serve as a fallback in case of unexpected or missed color usage.
    colors: {
        primary: 'var(--primary)',
        // Never used.
        primary75: 'var(--primary)',
        // Used for `option` background color, which is overwritten in `STYLES`.
        primary50: 'var(--primary-2)',
        // Used for `option` background color, which is overwritten in `STYLES`.
        primary25: 'var(--primary-2)',
        // Used for `multiValueRemove` color, which is replaced with a custom component.
        danger: 'var(--danger)',
        // Used for `multiValueRemove` background color, which is overwritten in `STYLES`.
        dangerLight: 'var(--danger)',
        // Used for `menu` and `control` background color as well as `option` color, all
        // of which are overwritten in `STYLES`.
        neutral0: 'var(--react-select-neutral0)',
        // Used for `control` background color and `placeholder` color, both of which are
        // overwritten in `STYLES`.
        neutral5: 'var(--react-select-neutral5)',
        // Used for `indicatorSeparator` background color and `control` border color, both
        // of which are overwritten in `STYLES`, and `multiValue` background color, which
        // is replaced with a custom component.
        neutral10: 'var(--react-select-neutral10)',
        // Used for `indicatorSeparator` background color, `control` border color, and
        // `option` color, all of which are overwritten in `STYLES`.
        neutral20: 'var(--react-select-neutral20)',
        // Used for `control` border color, which is overwritten in `STYLES`.
        neutral30: 'var(--react-select-neutral30)',
        // Used by components that aren't used for `isMulti=true` or have been replaced
        // with custom ones.
        neutral40: 'var(--react-select-neutral40)',
        // Used for `placeholder` color, which is overwritten in `STYLES`.
        neutral50: 'var(--react-select-neutral50)',
        // Used by components that have been replaced with custom ones.
        neutral60: 'var(--react-select-neutral60)',
        // Never used.
        neutral70: 'var(--react-select-neutral70)',
        // Used for `input` and `multiValue` color, which are both overwritten in
        // `STYLES`, as well as components that aren't used for `isMulti=true` or have
        // been replaced with custom ones.
        neutral80: 'var(--react-select-neutral80)',
        // Never used.
        neutral90: 'var(--react-select-neutral90)',
    },
})
