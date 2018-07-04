declare module 'deepmerge' {
    function deepmerge<T>(
        a: { [key: string]: T },
        b: { [key: string]: T },
        options: {
            arrayMerge: (a: T[], b: T[]) => T[]
        }
    ): { [key: string]: T }
    export = deepmerge
}
