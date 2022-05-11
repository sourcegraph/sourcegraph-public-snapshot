type ControlledAccessibleField<T> = Omit<T, 'aria-label' | 'id'>

type BaseFieldProps<T> = ControlledAccessibleField<T> & {
    className?: string
    /**
     * Used to control the styling of the field and surrounding elements.
     * Set this value to `false` to show invalid styling.
     * Set this value to `true` to show valid styling.
     */
    isValid?: boolean
    /**
     * Optional message to display below the form field.
     * This should typically be used to display additional information to the user.
     * It will be styled differently if `isValid` is truthy or falsy.
     */
    message?: React.ReactNode
}

type AssociatedLabelControlFieldProps<T> = BaseFieldProps<T> & {
    /**
     * Descriptive text rendered within a <label> element.
     */
    label: React.ReactNode
    /**
     * A unique ID for the field element. This is required to correctly associate the rendered <label> with the field.
     */
    id: string
}

type AriaLabelControlFieldProps<T> = BaseFieldProps<T> & {
    /**
     * Descriptive text that will be visually hidden but still read out by assistive technologies.
     */
    'aria-label': string
}

type AriaLabelledByControlFieldProps<T> = BaseFieldProps<T> & {
    /**
     * A valid ID to any element that can be used to describe to associated input.
     * This can be used for specific cases where we want fully control over the `label` element that is rendered.
     */
    'aria-labelledby': string
}

/**
 * This type should be used as a foundation for any form field elements within Wildcard.
 *
 * It enforces that we build accessible components through the following rules:
 * 1. Form fields should take an `isValid` prop to control styling for different form states.
 * 2. Form fields should take a `message` prop to display a message below the field.
 * 3. Form fields should provide descriptive text to ensure that assistive technologies are able to read the field.
 *
 * It enforces that accessible labels are attached to the field through either a `label` (and `id`) prop or an `aria-label` prop.
 */
export type AccessibleFieldProps<T> =
    | AssociatedLabelControlFieldProps<T>
    | AriaLabelControlFieldProps<T>
    | AriaLabelledByControlFieldProps<T>
