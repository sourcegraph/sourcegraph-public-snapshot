import { JSONSchema4 } from 'json-schema'
import refParser from 'json-schema-ref-parser'
import { cloneDeep, isEqual, without } from 'lodash'
import mapObj from 'map-obj'
import rawSchema from './lsif.schema.json'
import { Schema } from 'jsonschema'

/**
 * Loads the input and storage JSON schemas.
 */
export async function getJsonSchema(): Promise<Schema> {
    // Remove title attributes from $refs so they don't override the title of the referenced type
    // The title is used for prefixing predicates so it must always refer to the type name, not the property name
    // tslint:disable-next-line: no-any
    let modifiedSchema: Schema = mapObj<any, any, any>(
        rawSchema,
        (key, value, obj) => {
            if (value && value.$ref) {
                const { $ref } = value
                return [key, { $ref }]
            }
            return [key, value]
        },
        { deep: true }
    )

    // Rename for better predicate names:
    // Range_1 (LSIF vertex) -> Range
    // Range (LSP) -> lsp.Range
    // TODO add option to the schema generator to add namespace prefixes from imports automatically
    {
        const { Range, Range_1, ...rest } = modifiedSchema.definitions!
        modifiedSchema = {
            ...modifiedSchema,
            definitions: {
                ...rest,
                Range: Range_1,
                'lsp.Range': Range,
            },
        }
        // tslint:disable-next-line: no-any
        modifiedSchema = mapObj<any, any, any>(
            modifiedSchema,
            (key, value) =>
                key === '$ref'
                    ? [
                          key,
                          value
                              .replace(/^#\/definitions\/Range$/, '#/definitions/lsp.Range')
                              .replace(/^#\/definitions\/Range_1$/, '#/definitions/Range'),
                      ]
                    : [key, value],
            { deep: true }
        )
    }

    // We need two schemas:
    // One to validate the LSIF input
    // and one to validate the objects we store in the Graph DB,
    // which is slightly different.
    let storageSchema = cloneDeep(modifiedSchema)
    {
        // Replace Document.uri with Document.path relative file path
        // Remove contents
        const { uri, contents, ...documentRest } = storageSchema.definitions!.Document.properties!
        // Flatten ranges
        const { start, end, ...rangeRest } = storageSchema.definitions!.Range.properties!
        const flatRangeProperties: Schema['properties'] = {
            startLine: { type: 'integer' },
            startCharacter: { type: 'integer' },
            endLine: { type: 'integer' },
            endCharacter: { type: 'integer' },
        }
        storageSchema = {
            ...storageSchema,
            definitions: {
                ...storageSchema.definitions!,
                Document: {
                    ...storageSchema.definitions!.Document,
                    properties: {
                        ...documentRest,
                        path: { type: 'string' },
                    },
                    required: [...without(storageSchema.definitions!.Document.required as string[], 'uri')],
                },
                Range: {
                    ...storageSchema.definitions!.Range,
                    title: 'Range',
                    properties: {
                        ...rangeRest,
                        ...flatRangeProperties,
                    },
                    required: [
                        ...without(storageSchema.definitions!.Range.required as string[], 'start', 'end'),
                        ...Object.keys(flatRangeProperties),
                    ],
                },
                'lsp.Range': {
                    ...storageSchema.definitions!['lsp.Range'],
                    title: 'lsp.Range',
                },
            },
        }
    }

    return (await refParser.dereference(storageSchema as JSONSchema4)) as Schema
}

/** Returns the JSON schema type of a value */
function getTypeOf(value: unknown): string {
    if (value === null) {
        return 'null'
    }
    if (Array.isArray(value)) {
        return 'array'
    }
    if (typeof value === 'number' && Number.isInteger(value)) {
        return 'integer'
    }
    return typeof value
}

/**
 * An Error that is thrown by `validate()` when validation of a JSON schema failed.
 */
export class ValidationError extends Error {
    public readonly name: 'ValidationError'
    public readonly status: 422
    public schema: Schema
    public value: unknown
    public errors?: ValidationError[]

    constructor(message: string, schema: Schema, value: unknown, errors?: ValidationError[]) {
        if (schema.title) {
            super(`Schema "${schema.title}" did not match: ${message}`)
        } else {
            super(message)
        }
        this.name = 'ValidationError'
        this.schema = schema
        this.value = value
        if (errors) {
            this.errors = errors
        }
        this.status = 422
    }
}

/**
 * Validate the given value with the given schema and return the matched schema.
 * In the return value, any `anyOf` constraints will be replaced with the schema
 * that matched.
 *
 * NOTE: This does not fully implement the JSON schema spec, but only the subset
 * that typescript-json-schema produces.
 *
 * @throws {ValidationError} if the validation failed
 */
export function validate(schema: Schema, value: unknown): Schema {
    const validateWithContext = (s: Schema, v: unknown): Schema => {
        try {
            return validate(s, v)
        } catch (err) {
            if (!(err instanceof ValidationError)) {
                throw err
            }
            throw new ValidationError(err.message, schema, value, [err])
        }
    }
    if (schema.anyOf) {
        const errors: ValidationError[] = []
        for (const one of schema.anyOf) {
            try {
                return validateWithContext(one, value)
            } catch (err) {
                if (!(err instanceof ValidationError)) {
                    throw err
                }
                errors.push(err)
            }
        }
        throw new ValidationError('Expected anyOf to match', schema, value, errors)
    }
    const expectedType = schema.type || getTypeOf(value)
    const actualType = getTypeOf(value)
    if (Array.isArray(expectedType)) {
        for (const type of expectedType) {
            try {
                return validateWithContext({ ...schema, type }, value)
            } catch {
                // ignore
            }
        }
        throw new ValidationError('Expected type ' + expectedType.join(', '), schema, value)
    }
    if (expectedType !== actualType && !(expectedType === 'number' && actualType === 'integer')) {
        throw new ValidationError('Expected ' + schema.type, schema, value)
    }
    if (typeof value === 'number' && isNaN(value)) {
        throw new ValidationError('Value is NaN', schema, value)
    }
    if (schema.enum && !schema.enum.some(enumVal => isEqual(value, enumVal))) {
        throw new ValidationError(`Expected enum values ${schema.enum.join(', ')}, got ${value}`, schema, value)
    }
    if (expectedType === 'array') {
        if (!Array.isArray(value)) {
            throw new ValidationError('Expected array', schema, value)
        }
        const items: Schema[] = value.map((item, i) => {
            // Handle tuple typing
            if (schema.items) {
                if (Array.isArray(schema.items)) {
                    return validateWithContext(schema.items[i], item)
                }
                return validateWithContext(schema.items, item)
            }
            if (schema.additionalItems && typeof schema.additionalItems === 'object') {
                return validateWithContext(schema.additionalItems, item)
            }
            return {
                type: getTypeOf(item),
            }
        })
        if (Array.isArray(schema.items) && schema.additionalItems === false && value.length > schema.items.length) {
            throw new ValidationError('Additional array items not allowed', schema, value)
        }
        return { ...schema, items }
    }
    if (expectedType === 'object') {
        if (typeof value !== 'object' || value === null) {
            throw new ValidationError('Expected type object, got ' + actualType, schema, value)
        }
        const properties: { [k: string]: Schema } = {}
        for (const [key, propVal] of Object.entries(value)) {
            if (schema.properties && schema.properties[key]) {
                properties[key] = validateWithContext(schema.properties[key], propVal)
            } else if (schema.additionalProperties === false) {
                throw new ValidationError(`Unknown property ${key}`, schema, value)
            } else if (schema.additionalProperties && typeof schema.additionalProperties === 'object') {
                properties[key] = validateWithContext(schema.additionalProperties, propVal)
            } else {
                properties[key] = {
                    type: getTypeOf(propVal),
                }
            }
        }
        // Check all required keys are present
        if (schema.required) {
            const keys = new Set(Object.keys(value))
            const missingKeys = schema.required.filter(key => !keys.has(key))
            if (missingKeys.length > 0) {
                throw new ValidationError(`Missing required keys ` + missingKeys.join(', '), value, schema)
            }
        }
        return { ...schema, properties }
    }
    return schema
}
