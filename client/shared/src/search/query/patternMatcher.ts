/**
 * This module provides functions and TS types for pattern matching.
 * Given an input value and a pattern, `matchesValue` returns true if the value
 * is matched by the pattern.
 *
 * `PatternOf<...>` is the main TS type for a pattern.  What exactly can be used
 * as a pattern depends on the input value. For example If the input value is a
 * number value, then the pattern can be a function (returning a boolean) or a
 * number value. Things get a bit more complex when the input value is an object
 * or a union of objects.
 *
 * Here is a summary of possible pattern values:
 *
 * For primitive values, the pattern can be a value of the same type, a function
 * taking the value as argument and returning a boolean ("pattern function"), or
 * a "wrapper pattern" which is an object of the form '{$pattern: PatternOf<...>}'.
 * See 'WrapperPattern' below for more information.
 *
 *    matchesValue(42, 42)
 *    matchesValue(42, x => x > 0)
 *    matchesValue(42, {$pattern: 42})
 *
 * Additionally string values can be matched by a regular expression:
 *
 *    matchesValue('foo', /^f/)
 *    matchesValue('foo', {$pattern: /^f/})
 *
 * Arrays can only be matched by pattern functions.
 *
 *    matchesValue([1,2,3], values => values.every(x => x > 0))
 *
 * An object can be matched by an "object pattern", a pattern function or a
 * wrapper pattern.
 * The corresponding object pattern for an object has the same properties as the
 * input object, but they are all optional, and the values of those properties
 * are themselves patterns of the corresponding type.
 *
 *
 *    type O = {a: string; b: string|number}
 *    let o: O = {...}
 *    matchesValue(O, {a: 'foo'})
 *    matchesValue(O, {b: 42})
 *    matchesValue(O, {b: 'foo'})
 *    matchesValue(O, {b: x => typeof x === 'string'})
 *
 * If the input type is a union of objects, the properties of all members are
 * merged into into a single object. The type of every property is the union of
 * the respective property types, including 'undefined' if one of the objects in
 * the union doesn't have this property.
 *
 * Input type:
 *     {t: 'foo', a: string, b: number} | {t: 'bar', a: number}
 *
 * Pattern type:
 *     ObjectPattern<Input> = {
 *       t?: PatternFunction<'foo'|'bar'> | WrapperPattern<'foo'|'bar'> | 'foo' | 'bar',
 *       a?: PatternFunction<string|number> | WrapperPattern<string|number> | string | number
 *       b?: PatternFunction<string|undefined> | WrapperPattern<string|undefined> | string
 *     }
 *
 */

/**
 * ObjectPatterns and WrapperPattern can have an additional data annotation
 * ($data) to extract information if the corresponding pattern matches. The
 * extracted data is returned from `matchesValue`.
 */
interface DataAnnotation<Value, Data> {
    /**
     * This property allows you to extract/compute data from matched values. The
     * property can take a function and directly operate on context.data, or a
     * "mapping object". This data will be returned by `matchesValue`.
     *
     * Example:
     * matchesValue({a: 100}, {a: x => x > 50, $data: {outOfBounds: true}})
     *  => {
     *      success: true,
     *      data: {
     *        outOfBounds: true,
     *      }
     *    }
     */
    $data: DataMapper<Value, Data>
}

/**
 * This allows the data annotation to extract data statically or from the matched
 * value via a function.
 *
 * Example:
 * {
 *   $data: {
 *     static: 42,
 *     dynamic: value => value.x,
 *   }
 * }
 */
export type DataMapper<Value, Data> =
    | ((value: Value, context: MatchContext<Data>) => void)
    | (Data extends any[]
          ? never
          : { [K in keyof Data]?: ((value: Value, context: MatchContext<Data>) => Data[K]) | Data[K] })

/**
 * Creates the union of all keys of each member.
 *
 * Example: {a,c}|{a,b,d} => (a|c)|(a|b|d) => a|c|b|d
 */
type UnionKeys<T> = T extends any ? keyof T : never

/**
 * Given a union type and a property, this type alias will create the union of
 * property types from each union member, or 'undefined' if the property doesn't
 * exist.
 * Normally it's not possible to access a property on a union type that doesn't
 * exist in all of the members but because we operate on a generic type and we
 * extract all property names via 'UnionKeys', we can get away with it. However,
 * doing just '{[K in UnionKeys<T>]: T[K]}' will result in 'unknown' if one of
 * the members doesn't have the property.
 * With 'UnionValues', we get '... | undefined' instead.
 */
type UnionValues<Value, K> = Value extends any ? (K extends keyof Value ? Value[K] : undefined) : never

/**
 * A ObjectPattern matches the input if all pattern properties (which are a
 * subset of the input properties) match.
 *
 * Example: {foo: x => x > 0} matches {foo: 42}
 *
 * If the input type is a union of objects, the properties of all members are
 * merged into into a single object. The type of every property is the union of
 * the respective property types, including 'undefined' if one of the objects in
 * the union doesn't have this property.
 *
 * Example:
 * Input:
 *     {t: 'foo', a: string, b: number} | {t: 'bar', a: number}
 *
 * Output:
 *     ObjectPattern<Input> = {
 *       t?: PatternFunction<'foo'|'bar'> | WrapperPattern<'foo'|'bar'> | 'foo' | 'bar',
 *       a?: PatternFunction<string|number> | WrapperPattern<string|number> | string | number
 *       b?: PatternFunction<string|undefined> | WrapperPattern<string|undefined> | string
 *     }
 */
type ObjectPattern<Value, Data> =
    // ObjectPattern is applied below with ObjectPattern<ObjectMembers<Value>>. The
    // check against 'never' prevents empty unions.
    [Value] extends [never]
        ? never
        : { [K in UnionKeys<Value>]?: PatternOf<UnionValues<Value, K>, Data> } & Partial<WrapperPattern<Value, Data>>

/**
 * A WrapperPattern makes it possible to add data annotations to patterns that
 * are not object patterns. In the following example, it wouldn't be possible to
 * attach additional data to the different values of "type":
 *
 * {type: oneOf("script", "module")}
 *
 * We could write
 *
 * {type: oneOf("script", "module"), $: {module: value => value.type === "module"}
 *
 * but in order to avoid precedural logic as much as possible, WrapperPattern
 * allows us to write
 *
 * {type: oneOf("script", {$pattern: "module", $: {module: true}})}
 */
interface WrapperPattern<Value, Data> extends DataAnnotation<Value, Data> {
    /**
     * The pattern to match the input value against.
     */
    $pattern: PatternFunction<Value, Data> | PrimitivePattern<Value>
}

function isWrapperPattern(value: any): value is WrapperPattern<any, any> {
    return typeof value === 'object' && '$pattern' in value
}

/**
 * A PatternFunction accepts the value to match against, the current match
 * context (which also holds the extracted data) and the internal match function
 * for convenience. See the combinatorial helper functions below for examples.
 */
export type PatternFunction<Value, Data> = (value: Value, context: MatchContext<Data>, match: typeof matches) => boolean

/**
 * This type is used for matching against primitives. The main reason for its
 * existence is to allow first class support for matching strings against
 * regular expressions.
 */
type PrimitivePattern<Value> = Value extends string
    ? RegExp | Value
    : Value extends number | boolean | null | undefined
    ? Value
    : never

/**
 * Removes all non-object members from Value
 */
type ObjectMembers<Value> = Value extends object ? Value : never

/**
 * The main pattern type. What a valid pattern for a value is depends on the
 * value:
 * - For an array, the pattern has to be a function
 * - For an object, the pattern can be an object pattern, a wrapper pattern or a
 *   pattern function
 * - For a primitive value, the pattern can be a value of the same type, a
 *   wrapper pattern or a pattern function. For strings the pattern can also be
 *   a regular expression.
 */
export type PatternOf<Value, Data = unknown> =
    // A pattern function is always a valid value. The function receives the
    // value to be matched as argument.
    | PatternFunction<Value, Data>
    | WrapperPattern<Value, Data> // Arrays always have to matched with a pattern function
    // The [...] around the types are necessary to avoid
    // distributing union types.
    | ([Value] extends [any[]]
          ? never
          : // Note that we are not checking types here (Value extends ...). For
            // one, union types of mixed types (e.g. number|{x: number}) makes
            // this a bit annoying, and what we really want to do do here is
            // _filter_ 'Value' and allow/create an ObjectPattern for the objects in
            // the union, and PrimitivePatterns for the primtives.
            // ObjectMembers does the object filtering and PrimitivePattern
            // includes type checks itself.
            ObjectPattern<ObjectMembers<Value>, Data> | PrimitivePattern<Value>)

/**
 * Utility type to prevent TS from inferring the type from a function parameter
 *
 * @see https://stackoverflow.com/questions/56687668
 */
type NoInfer<A extends any> = [A][A extends any ? 0 : never]

/**
 * Helper type to indicate to infer the values for Value, Data from other input
 * and not from the pattern itself.
 */
export type PatternOfNoInfer<Value, Data> = PatternOf<NoInfer<Value>, NoInfer<Data>>

/**
 * Context that is passed to all pattern functions during a single pattern
 * application. Currently only holds the extracted data.
 */
export interface MatchContext<Data> {
    data: Data
}

type Result<Data> =
    | {
          success: true
          /**
           * The data extract from the input value by the pattern.
           */
          data: Data
      }
    | { success: false }

/**
 * "Applies" the 'pattern' to 'value'. It returns whether the pattern matched
 * and all data extracted by patterns.
 * If 'Data' is an object all of its properties should be optional.
 */
export function matchesValue<Value, Data>(
    value: Value,
    pattern: PatternOfNoInfer<Value, Data>,
    initialData: Data = Object.create(null)
): Result<Data> {
    const context = { data: initialData ?? Object.create(null) }
    if (matches(context, value, pattern)) {
        return { success: true, data: context.data ?? initialData }
    }
    return { success: false }
}

const specialKeys: Set<string> = new Set(['$pattern', '$data'])

/**
 * Internal match function that does the heavy lifting and is passed to every
 * pattern function.
 */
function matches<Value, Data>(
    context: MatchContext<Data>,
    value: Value,
    pattern: PatternOfNoInfer<Value, Data> | undefined
): boolean {
    if (typeof pattern === 'function') {
        const result = pattern(value, context, matches)
        return result
    }
    if (pattern instanceof RegExp) {
        return typeof value === 'string' && pattern.test(value)
    }
    if (pattern && typeof pattern === 'object') {
        let match = false
        let matchKeys = value && typeof value === 'object'

        if (isWrapperPattern(pattern)) {
            // TS2345:  Type 'RegExp' is not assignable to type '[NoInfer<Value>] extends [any[]] ? never : WrapperPattern<NoInfer<Value>, NoInfer<Data>> | ObjectPattern<ObjectMembers<NoInfer<Value>>, NoInfer<Data>> | PrimitivePattern<NoInfer<Value>>'
            // @ts-expect-error TBH I don't know why RegExp is causing problems here
            match = matches(context, value, pattern.$pattern)
            // There is no point in looking at the remaining properties if the
            // $pattern function didn't match.
            matchKeys = match
        }
        if (matchKeys) {
            const keys = Object.getOwnPropertyNames(pattern).filter(key => !specialKeys.has(key))
            if (keys.length > 0) {
                match = keys.every(
                    // TS7053: Element implicitly has an 'any' type because expression of type 'string' can't be used to index type 'unknown'.  No index signature with a parameter of type 'string' was found on type 'unknown'.
                    // TS7053: Element implicitly has an 'any' type because expression of type 'string' can't be used to index type 'WrapperPattern<NoInfer<Value>, NoInfer<Data>> | ...
                    // @ts-expect-error this will properly always error because 'value' and 'pattern' are normally not indexable
                    property => property in value && matches(context, value[property], pattern[property])
                )
            }
        }
        // else: either no match or value isn't an object (but pattern is) so there can't be a match

        // Special property to capture values
        if (match && pattern.$data) {
            if (typeof pattern.$data === 'function') {
                //  TS2345: Argument of type 'Value' is not assignable to parameter of type 'ObjectMembers<NoInfer<Value>>'
                // @ts-expect-error due to the type definition above, pattern.$data is a union of two functions
                pattern.$data(value, context)
            } else {
                for (const [key, captureValue] of Object.entries(pattern.$data)) {
                    // TS7053: Element implicitly has an 'any' type because expression of type 'string' can't be used to index type 'unknown'.  No index signature with a parameter of type 'string' was found on type 'unknown'
                    // @ts-expect-error context.data will likely not be an indexable type
                    context.data[key] =
                        typeof captureValue === 'function' ? captureValue(value, context.data) : captureValue
                }
            }
        }
        return match
    }
    return value === pattern
}

// Standard pattern functions

// Combinators

/**
 * Matches if any of its elements match the provided pattern.
 */
export function some<Value, Data>(pattern: PatternOfNoInfer<Value, Data>): PatternFunction<Value[], Data> {
    return (values, context, matches) => values.some(value => matches(context, value, pattern))
}

/**
 * Matches if all of the elements match the provided pattern.
 */
export function every<Value, Data>(pattern: PatternOfNoInfer<Value, Data>): PatternFunction<Value[], Data> {
    return (values, context, matches) => values.every(value => matches(context, value, pattern))
}

/**
 * Always matches. Its main purpose is to apply a pattern to multiple input
 * values to extract data from each of them.
 */
export function each<Value, Data>(pattern: PatternOfNoInfer<Value, Data>): PatternFunction<Value[], Data> {
    return (values, context, matches): true => {
        for (const value of values) {
            matches(context, value, pattern)
        }
        return true
    }
}

/**
 * Matches if the value matches one of the patterns.
 */
export function oneOf<Value, Data>(...patterns: PatternOfNoInfer<Value, Data>[]): PatternFunction<Value, Data> {
    return (value, context, matches) => patterns.some(pattern => matches(context, value, pattern))
}

/**
 * Matches if the value matches all of the patterns.
 */
export function allOf<Value, Data>(...patterns: PatternOfNoInfer<Value, Data>[]): PatternFunction<Value, Data> {
    return (value, context, matches) => patterns.every(pattern => matches(context, value, pattern))
}

/**
 * Similar to each above but matching multiple patterns against a single value
 * to extract information. Always returns true.
 */
export function eachOf<Value, Data>(...patterns: PatternOfNoInfer<Value, Data>[]): PatternFunction<Value, Data> {
    return (value, context, matches): true => {
        for (const pattern of patterns) {
            matches(context, value, pattern)
        }
        return true
    }
}

/**
 * Matches if the value does not match the pattern.
 */
export function not<Value, Data>(pattern: PatternOfNoInfer<Value, Data>): PatternFunction<Value, Data> {
    return (value, context, matches) => !matches(context, value, pattern)
}

// Misc

/**
 * Helper function to debug patterns
 */
export function debug<Value, Data>(pattern: PatternOfNoInfer<Value, Data>): PatternFunction<Value, Data> {
    return (value, context, matches) => {
        // eslint-disable-next-line no-debugger
        debugger
        // These are intentially two statements to make it easier to inspect the
        // result of calling `matches` in the debugger.
        const result = matches(context, value, pattern)
        return result
    }
}
