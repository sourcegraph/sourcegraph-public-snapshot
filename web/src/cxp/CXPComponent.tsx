import { Component } from 'cxp/module/environment/environment'
import * as React from 'react'

/** React props for components that participate in the CXP environment. */
export interface CXPComponentProps {
    /**
     * Called when the CXP component changes.
     */
    cxpOnComponentChange: (component: Component | null) => void
}

interface Props extends CXPComponentProps {
    /** A description of the parent component (e.g., Blob) that is presenting the document. */
    component: Component | null
}

/**
 * A component that participates in CXP. A participating React component's render method should include a
 * <CXPComponent> element that describes the React component's state.
 */
export class CXPComponent extends React.PureComponent<Props> {
    public componentDidMount(): void {
        this.updateEnvironment()
    }

    public componentDidUpdate(): void {
        this.updateEnvironment()
    }

    public componentWillUnmount(): void {
        this.props.cxpOnComponentChange(null)
    }

    public render(): JSX.Element | null {
        return null
    }

    private updateEnvironment(): void {
        this.props.cxpOnComponentChange(this.props.component)
    }
}
