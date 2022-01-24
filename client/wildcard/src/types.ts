import * as React from 'react'

type Merge<P1 = {}, P2 = {}> = Omit<P1, keyof P2> & P2

export type ForwardReferenceExoticComponent<E, OwnProps> = React.ForwardRefExoticComponent<
    Merge<E extends React.ElementType ? React.ComponentPropsWithRef<E> : never, OwnProps & { as?: E }>
>

export interface ForwardReferenceComponent<
    IntrinsicElementString,
    OwnProps = {}
    /**
     * Extends original type to ensure built in React types play nice
     * with polymorphic components still e.g. `React.ElementRef` etc.
     */
> extends ForwardReferenceExoticComponent<IntrinsicElementString, OwnProps> {
    /**
     * When `as` prop is passed, use this overload.
     * Merges original own props (without DOM props) and the inferred props
     * from `as` element with the own props taking precedence.
     *
     * The exception is made for the `ref` prop that changes based on the `as` prop value.
     * Inferred `ref` props overwrites the original `ref` type.
     *
     * We explicitly avoid `React.ElementType` and manually narrow the prop types
     * so that events are typed when using JSX.IntrinsicElements.
     */
    <As = IntrinsicElementString, OwnPropsOverwrite = OwnProps>(
        props: As extends ''
            ? { as: keyof JSX.IntrinsicElements }
            : As extends React.ComponentType<infer P>
            ? Merge<P, Omit<OwnPropsOverwrite, 'ref'> & { as: As }>
            : As extends keyof JSX.IntrinsicElements
            ? Merge<JSX.IntrinsicElements[As], Omit<OwnPropsOverwrite, 'ref'> & { as: As }>
            : never
    ): React.ReactElement | null
}
