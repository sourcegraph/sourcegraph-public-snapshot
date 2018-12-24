// We can't use @types/react-visibility-sensor because its type for the component's `children` prop
// doesn't support HTML elements (like <code> as we use), only React components and child functions.

import * as React from 'react'

export as namespace ReactVisibilitySensor

interface Shape {
    top?: number
    left?: number
    bottom?: number
    right?: number
}

interface ChildFunctionArg {
    isVisible: boolean | null
    visibilityRect: Shape
}

type ChildFunction = (arg: ChildFunctionArg) => React.ReactNode

interface Props {
    onChange?: (isVisible: boolean) => void
    active?: boolean
    partialVisibility?: boolean
    offset?: Shape
    minTopValue?: number
    intervalCheck?: boolean
    intervalDelay?: number
    scrollCheck?: boolean
    scrollDelay?: number
    scrollThrottle?: number
    resizeCheck?: boolean
    resizeDelay?: number
    resizeThrottle?: number
    containment?: HTMLElement
    delayedCall?: boolean
    children?: React.ReactNode | ChildFunction // modified from @types/react-visibility-sensor
}

declare const ReactVisibilitySensor: React.FunctionComponent<Props>

declare module 'react-visibility-sensor' {
    export = ReactVisibilitySensor
}
