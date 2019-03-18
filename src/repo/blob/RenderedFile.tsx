import * as H from 'history'
import * as React from 'react'
import { Markdown } from '../../components/Markdown'

interface Props {
    /**
     * The rendered HTML contents of the file.
     */
    dangerousInnerHTML: string

    location: H.Location
}

/**
 * Displays a file whose contents are rendered to HTML, such as a Markdown file.
 */
export class RenderedFile extends React.PureComponent<Props> {
    public componentDidMount(): void {
        if (this.props.dangerousInnerHTML && this.props.location.hash) {
            this.scrollToHash(this.props.location.hash)
        }
    }

    public componentDidUpdate(prevProps: Props): void {
        if (
            prevProps.dangerousInnerHTML !== this.props.dangerousInnerHTML ||
            prevProps.location.hash !== this.props.location.hash
        ) {
            // Try scrolling when either the content or the hash changed.
            this.scrollToHash(this.props.location.hash)
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className="rendered-file">
                <div className="rendered-file__container">
                    <Markdown dangerousInnerHTML={this.props.dangerousInnerHTML} />
                </div>
            </div>
        )
    }

    /** Scroll to the anchor in the page identified by the location fragment (e.g., #my-section). */
    private scrollToHash(hash: string): void {
        if (!hash) {
            return
        }

        // This does not cause a page navigation (because window.location.hash === hash already), but it does cause
        // the page to scroll to the hash. This is simpler than using scrollTo, scrollIntoView, etc. Also assigning
        // window.location.hash does not trigger a navigation when `window.location.hash === hash`, so we can't
        // just use that.
        //
        // Finally, ensure that hash begins with a "#" so that this can't be used to redirect to arbitrary URLs (as
        // an open redirect vulnerability). This should always be true, but just be safe and document this
        // assertion in code.
        if (hash.startsWith('#')) {
            window.location.href = hash
        }
    }
}
