import { type FC, cloneElement, type ReactElement, type ReactNode } from 'react'

interface ComponentsComposerProps {
    components: ReactElement[]
    children: ReactNode[]
}

/**
 * Composes an array of components into a nested tree.
 * Used to simplify our render methods by composing multiple context providers together.
 *
 * ```tsx
 *   <ComponentsComposer contexts={[
 *      <WildcardThemeContext.Provider key="wildcard" value={WILDCARD_THEME} />,
 *      <ErrorBoundary key="error-boundary" location={null} />,
 *   ]}>
 *       <main>My app</main>
 *       <footer />
 *   </ComponentsComposer>
 * ```
 *
 */
export const ComponentsComposer: FC<ComponentsComposerProps> = ({ components, children }) =>
    components.reduceRight(
        (children, parent) =>
            cloneElement(parent, {
                children,
            }),
        <>{children}</>
    )
