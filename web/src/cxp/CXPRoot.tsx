import { URI } from 'cxp/module/types/textDocument'
import * as React from 'react'

/** React props for components that participate in the creation or lifecycle of a CXP root. */
export interface CXPRootProps {
    /**
     * Called when the CXP root changes.
     */
    cxpOnRootChange: (root: URI | null) => void
}

interface Props extends CXPRootProps {
    /** The root URI that the parent component (e.g., RepoRevContainer) represents. */
    root: URI | null
}

/**
 * A component that participates in the creation or lifecycle of a CXP root. A participating React component's
 * render method should include a <CXPRoot> element that describes the root represented by the React component.
 */
export class CXPRoot extends React.PureComponent<Props> {
    public componentDidMount(): void {
        this.updateEnvironment()
    }

    public componentDidUpdate(): void {
        this.updateEnvironment()
    }

    public componentWillUnmount(): void {
        this.props.cxpOnRootChange(null)
    }

    public render(): JSX.Element | null {
        return null
    }

    private updateEnvironment(): void {
        this.props.cxpOnRootChange(this.props.root)
    }
}
