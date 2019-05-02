declare module 'simmerjs' {
    export interface Options {
        /**
         * @default 100
         */
        specificityThreshold?: number

        /**
         * @default 3
         */
        depth?: number

        /**
         * @default 520
         */
        selectorMaxLength?: number

        errorHandling?: boolean | ((error: any, element: Element) => void)
    }

    export type QueryEngine = (selector: string, onError: (error: any) => void) => ArrayLike<Element>
    export interface Queryable {
        querySelectorAll: QueryEngine
    }

    export interface WindowLike {
        document: Queryable
    }
    export type Scope = Queryable | WindowLike

    export type Simmer = (element: Element) => string
    interface SimmerConstructor {
        new (scope?: Scope, options?: Options, query?: QueryEngine): Simmer
        (scope?: Scope, options?: Options, query?: QueryEngine): Simmer

        configure(options?: Options): void
        noConflict(): SimmerConstructor
    }
    const Simmer: SimmerConstructor
    export default Simmer
}
