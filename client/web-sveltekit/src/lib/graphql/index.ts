export * from './apollo'

// Helper type for extracting node() query related type information
export type NodeFromResult<T extends { __typename: string } | null, N extends string> = T extends { __typename: N }
    ? NonNullable<T>
    : never
