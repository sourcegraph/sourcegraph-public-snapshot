import { URI } from '@sourcegraph/sourcegraph.proposed/module/types/textDocument'
import * as React from 'react'

/** React props for components that participate in the creation or lifecycle of a CXP root. */
export interface ExtensionsRootProps {
    /**
     * Called when the CXP root changes.
     */
    extensionsOnRootChange: (root: URI | null) => void
}

interface Props extends ExtensionsRootProps {
    /** The root URI that the parent component (e.g., RepoRevContainer) represents. */
    root: URI | null
}

/**
 * A component that participates in the creation or lifecycle of a CXP root. A participating React component's
 * render method should include a <ExtensionsRoot> element that describes the root represented by the React component.
 */
export class ExtensionsRoot extends React.PureComponent<Props> {
    public componentDidMount(): void {
        this.updateEnvironment()
    }

    public componentDidUpdate(): void {
        this.updateEnvironment()
    }

    public componentWillUnmount(): void {
        this.props.extensionsOnRootChange(null)
    }

    public render(): JSX.Element | null {
        return null
    }

    private updateEnvironment(): void {
        this.props.extensionsOnRootChange(this.props.root)
    }
}
