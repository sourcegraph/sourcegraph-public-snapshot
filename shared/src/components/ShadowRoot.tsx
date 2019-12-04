import * as React from 'react'
import * as ReactDOM from 'react-dom'

/**
 * Possible elements to function as a shadow root host
 */
export type HostElement =
    | 'article'
    | 'aside'
    | 'blockquote'
    | 'body'
    | 'div'
    | 'footer'
    | 'h1'
    | 'h2'
    | 'h3'
    | 'h4'
    | 'h5'
    | 'h6'
    | 'header'
    | 'main'
    | 'nav'
    | 'p'
    | 'section'
    | 'span'

interface HostElementProp<H extends HostElement> {
    /**
     * Element type of the host element
     *
     * @default 'div'
     */
    hostElement: H
}

export type ShadowRootProps<H extends HostElement> = JSX.IntrinsicElements[H] &
    (H extends 'div' ? Partial<HostElementProp<H>> : HostElementProp<H>) & {
        /**
         * CSS to be added to the shadow root (through a `<style>` tag).
         * Import this from a .scss file like this:
         *
         * ```tsx
         * import myComponentStyles from './MyComponent.scss'
         *
         * const MyComponent = () => <ShadowRoot styles={[myComponentStyles]}>Hello World!</ShadowRoot>
         * ```
         */
        css?: (string | { toString(): string })[]
    }

/**
 * A wrapper component that renders its children in a shadow DOM.
 *
 * To learn more about shadow roots, see https://developers.google.com/web/fundamentals/web-components/shadowdom
 */
export class ShadowRoot<H extends HostElement> extends React.Component<ShadowRootProps<H>, {}> {
    private hostRef = React.createRef<Element>()

    public componentDidMount(): void {
        this.componentDidUpdate()
    }

    public componentDidUpdate(): void {
        if (!this.hostRef.current) {
            return
        }
        if (!this.hostRef.current.shadowRoot) {
            this.hostRef.current.attachShadow({ mode: 'open' })
        }
    }

    public render(): React.ReactNode {
        const { hostElement: Host = 'div' as any, css, children, ...hostProps } = this.props
        return (
            <>
                <Host ref={this.hostRef} {...hostProps} />
                {this.hostRef.current &&
                    this.hostRef.current.shadowRoot &&
                    ReactDOM.createPortal(
                        <>
                            {css?.map((css, i) => (
                                <style key={i}>{css.toString()}</style>
                            ))}
                            {children}
                        </>,
                        this.hostRef.current.shadowRoot as any
                    )}
            </>
        )
    }
}
