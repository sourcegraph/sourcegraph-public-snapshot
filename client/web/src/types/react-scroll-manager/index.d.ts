/* eslint-disable react/prefer-stateless-function */
import React from 'react'

import { History } from 'history'

export interface ScrollManagerProps {
    history: History
    sessionKey?: string
    timeout?: number
}

export class ScrollManager extends React.Component<React.PropsWithChildrenq<ScrollManagerProps>> {}

export class WindowScroller extends React.Component {}

export interface ElementScrollerProps {
    scrollKey: string
}

export class ElementScroller extends React.Component<React.PropsWithChildrenq<ElementScrollerProps>> {}
