import { Attributes } from '@opentelemetry/api'

export function parseAttributes(attributes: Attributes): Attributes {
    return Object.fromEntries(
        Object.entries(attributes).map(([key, value]) => {
            try {
                return [key, JSON.parse(value as string)]
            } catch {
                return [key, value]
            }
        })
    )
}
