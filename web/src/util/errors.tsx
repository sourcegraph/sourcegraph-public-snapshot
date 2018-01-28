export interface ErrorLike {
    message?: string
    code?: string
}

export const isErrorLike = (val: any): val is { message?: string; code?: string } =>
    val && typeof val === 'object' && ('message' in val || 'code' in val)
