// Error is a type that describes errors returned by XyzBackends if a network
// request fails. Generally you will use Error in a union like `MyType | Error`,
// so that you force callers to handle errors.
export type Error = {Error: any};
