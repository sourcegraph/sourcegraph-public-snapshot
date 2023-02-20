import { lazy, Attributes, ComponentType, FC, PropsWithRef } from 'react'

type ComponentFactory<P extends {}, K extends string> = () => Promise<{ [key in K]: ComponentType<P> }>

/**
 * Returns a lazy-loaded reference to a React component in another module.
 *
 * This should be used in URL routes and anywhere else that Webpack code splitting can occur, to
 * avoid all referenced components being in the initial bundle.
 *
 * @param componentFactory Asynchronously imports the component's module; e.g., `() =>
 * import('./MyComponent')`.
 * @param name The export binding name of the component in its module.
 */
export const lazyComponent = <Props extends {}, Name extends string>(
    componentFactory: ComponentFactory<Props, Name>,
    name: Name
): FC<PropsWithRef<Props> & Attributes> => {
    // Force returning a FunctionComponent-like so our result is callable (because it's used
    // in <Route render={...} /> elements where it is expected to be callable).
    const LazyComponent = lazy(async () => {
        const component: ComponentType<Props> = (await componentFactory())[name]

        return { default: component }
    })

    return props => <LazyComponent {...props} />
}

export const lazyStormComponent = <Props extends {}, Name extends string>(
    componentFactory: ComponentFactory<Props, Name>,
    name: Name
): FC<PropsWithRef<Props> & Attributes> => {
    // Force returning a FunctionComponent-like so our result is callable (because it's used
    // in <Route render={...} /> elements where it is expected to be callable).
    const LazyComponent = lazy(async () => {
        const component: ComponentType<Props> = (await componentFactory())[name]

        return { default: component }
    })

    return LazyComponent
}

export const lazyRoute = <Props extends {}, Name extends string>(
    componentFactory: ComponentFactory<Props, Name>,
    name: Name,
    loaderFactory: any
) => {
    let LazyActual = lazyComponent<Props, Name>(componentFactory, name)

    async function lazyLoader(...args: any[]) {
        let controller = new AbortController()

        /*
         * Kick off our component chunk load but don't await it
         * This allows us to parallelize the component download with loader
         * download and execution.
         *
         * Normal React.lazy()
         *
         *   load loader.ts     execute loader()   load component.ts
         *   -----------------> -----------------> ----------------->
         *
         * Kicking off the component load _in_ your loader()
         *
         *   load loader.ts     execute loader()
         *   -----------------> ----------------->
         *                      load component.ts
         *                      ----------------->
         *
         * Kicking off the component load _alongside_ your loader.ts chunk load
         *
         *   load loader.ts     execute loader()
         *   -----------------> ----------------->
         *   load component.ts
         *   ----------------->
         */
        componentFactory().then(
            componentModule => {
                if (!controller.signal.aborted) {
                    // We loaded the component _before_ our loader finished, so we can
                    // blow away React.lazy and just use the component directly.  This
                    // avoids the flicker we'd otherwise get since React.lazy would need
                    // to throw the already-resolved promise up to the Suspense boundary
                    // one time to get the resolved value
                    LazyActual = componentModule as any
                }
            },
            () => {}
        )

        try {
            // Load our loader chunk
            let { default: loader } = await loaderFactory()
            // Call the loader
            return await loader(...args)
        } finally {
            // Abort the controller when our loader finishes.  If we finish before the
            // component chunk loads, this will ensure we still use React.lazy to
            // render the component since it's not yet available.  If the component
            // chunk finishes first, it will have overwritten Lazy with the legit
            // component so we'll never see the suspense fallback
            controller.abort()
        }
    }

    return {
        element: <LazyActual />,
        loader: lazyLoader,
    }
}
