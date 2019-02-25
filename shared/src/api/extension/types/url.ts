export const isURL = (value: any): value is URL =>
    !!value && typeof value.toString === 'function' && value.href === value.toString()
