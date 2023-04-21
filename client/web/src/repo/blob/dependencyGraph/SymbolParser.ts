export enum Descriptor_Suffix {
    UnspecifiedSuffix = 0,
    Namespace = 1,
    Package = 1,
    Type = 2,
    Term = 3,
    Method = 4,
    TypeParameter = 5,
    Parameter = 6,
    Meta = 7,
    Local = 8,
    Macro = 9,
}

type Descriptor = {
    name: string
    disambiguator: string
    suffix: Descriptor_Suffix
}

interface Package {
    manager: string
    name: string
    version: string
}

export interface ISymbol {
    scheme: string
    package?: Package
    descriptors: Descriptor[]
}

export interface SCIPDocument {
    relative_path: string
    occurrences: { symbol: string }[]
    symbols: { symbol: string }[]
}

// export interface SCIPOccurrence {
//     range: number[]
//     symbol: string
//     symbol_roles: number
//     syntax_kind: number
// }

function isSpecialCharacter(r: string): boolean {
    return /[\(\)\[\]\/\.\#\:!]/.test(r)
}

function isIdentifierCharacter(r: string): boolean {
    return /[a-zA-Z0-9\-+$_]/.test(r)
}

class SymbolParser {
    symbolString: string
    symbol: string[]
    index: number

    constructor(symbol: string) {
        this.symbolString = symbol
        this.symbol = Array.from(symbol)
        this.index = 0
    }

    current(): string {
        return this.symbol[this.index]
    }

    acceptSpaceEscapedIdentifier(what: string): string | Error {
        return this.acceptEscapedIdentifier(what, ' ')
    }

    acceptEscapedIdentifier(what: string, escapeCharacter: string): string | Error {
        let builder: string[] = []
        while (this.index < this.symbol.length) {
            const ch = this.current()
            if (ch === escapeCharacter) {
                this.index++
                if (this.index >= this.symbol.length) {
                    break
                }
                if (this.current() === escapeCharacter) {
                    // Escaped space character.
                    builder.push(ch)
                } else {
                    return builder.join('')
                }
            } else {
                builder.push(ch)
            }
            this.index++
        }
        return new Error(`reached end of symbol while parsing <${what}>, expected a '${escapeCharacter}' character`)
    }

    acceptBacktickEscapedIdentifier(what: string): string | Error {
        return this.acceptEscapedIdentifier(what, '`')
    }

    error(message: string): Error {
        return new Error(`${message}\n${this.symbolString}\n${'_'.repeat(this.index)}^`)
    }

    acceptIdentifier(what: string): string | Error {
        if (this.current() === '`') {
            this.index++
            return this.acceptBacktickEscapedIdentifier(what)
        }
        const start = this.index
        while (this.index < this.symbol.length && isIdentifierCharacter(this.current())) {
            this.index++
        }
        if (start === this.index && !isSpecialCharacter(this.current())) {
            return this.error('empty identifier')
        }
        return this.symbol.slice(start, this.index).join('')
    }

    acceptCharacter(r: string, what: string): Error | null {
        if (this.current() === r) {
            this.index++
            return null
        }
        return this.error(`expected '${r}', obtained '${this.current()}', while parsing ${what}`)
    }

    peekNext(): string | null {
        if (this.index + 1 < this.symbol.length) {
            return this.symbol[this.index]
        }
        return null
    }

    parseDescriptors(): Descriptor[] | Error {
        const result: Descriptor[] = []
        while (this.index < this.symbol.length) {
            const descriptorResult = this.parseDescriptor()
            if (descriptorResult instanceof Error) {
                return descriptorResult
            }
            const descriptor = descriptorResult
            result.push(descriptor)
        }
        return result
    }

    parseDescriptor(): Descriptor | Error {
        switch (this.peekNext()) {
            case '(':
                this.index++
                const paramName = this.acceptIdentifier('parameter name')
                if (paramName instanceof Error) {
                    return paramName
                }
                const closingParamError = this.acceptCharacter(')', 'closing parameter name')
                if (closingParamError) {
                    return closingParamError
                }
                return {
                    name: paramName,
                    disambiguator: '',
                    suffix: Descriptor_Suffix.Parameter,
                }
            case '[':
                this.index++
                const typeParamName = this.acceptIdentifier('type parameter name')
                if (typeParamName instanceof Error) {
                    return typeParamName
                }
                const closingTypeParamError = this.acceptCharacter(']', 'closing type parameter name')
                if (closingTypeParamError) {
                    return closingTypeParamError
                }
                return {
                    name: typeParamName,
                    disambiguator: '',
                    suffix: Descriptor_Suffix.TypeParameter,
                }
            default:
                const descName = this.acceptIdentifier('descriptor name')
                if (descName instanceof Error) {
                    return descName
                }
                const suffix = this.current()
                this.index++
                switch (suffix) {
                    case '(':
                        let disambiguator: string | Error = ''
                        if (this.peekNext() !== ')') {
                            disambiguator = this.acceptIdentifier('method disambiguator')
                            if (disambiguator instanceof Error) {
                                return disambiguator
                            }
                        }
                        const closingMethodError = this.acceptCharacter(')', 'closing method')
                        if (closingMethodError) {
                            return closingMethodError
                        }
                        return {
                            name: descName,
                            disambiguator,
                            suffix: Descriptor_Suffix.Method,
                        }
                    case '/':
                        return {
                            name: descName,
                            disambiguator: '',
                            suffix: Descriptor_Suffix.Namespace,
                        }
                    case '.':
                        return {
                            name: descName,
                            disambiguator: '',
                            suffix: Descriptor_Suffix.Term,
                        }
                    case '#':
                        return {
                            name: descName,
                            disambiguator: '',
                            suffix: Descriptor_Suffix.Type,
                        }
                    case ':':
                        return {
                            name: descName,
                            disambiguator: '',
                            suffix: Descriptor_Suffix.Meta,
                        }
                    case '!':
                        return {
                            name: descName,
                            disambiguator: '',
                            suffix: Descriptor_Suffix.Macro,
                        }
                    default:
                    // Handle the default case here...
                }
        }
        return new Error('Unable to parse descriptor') // Or return an appropriate default value or error
    }
}

export function parseSymbol(symbol: string): ISymbol | Error {
    return parsePartialSymbol(symbol, true)
}

function newLocalSymbol(id: string): ISymbol {
    return {
        scheme: 'local',
        descriptors: [
            {
                name: id,
                disambiguator: '',
                suffix: Descriptor_Suffix.Local,
            },
        ],
    }
}

// The ParsePartialSymbol function
function parsePartialSymbol(symbol: string, includeDescriptors: boolean): ISymbol | Error {
    const s = new SymbolParser(symbol)
    const schemeResult = s.acceptSpaceEscapedIdentifier('scheme')

    if (schemeResult instanceof Error) {
        return schemeResult
    }

    const scheme = schemeResult

    if (scheme === 'local') {
        return newLocalSymbol(s.symbol.slice(s.index).join(''))
    }

    const managerResult = s.acceptSpaceEscapedIdentifier('package manager')

    if (managerResult instanceof Error) {
        return managerResult
    }

    let manager = managerResult

    if (manager === '.') {
        manager = ''
    }

    const packageNameResult = s.acceptSpaceEscapedIdentifier('package name')

    if (packageNameResult instanceof Error) {
        return packageNameResult
    }

    let packageName = packageNameResult

    if (packageName === '.') {
        packageName = ''
    }

    const packageVersionResult = s.acceptSpaceEscapedIdentifier('package version')

    if (packageVersionResult instanceof Error) {
        return packageVersionResult
    }

    let packageVersion = packageVersionResult

    if (packageVersion === '.') {
        packageVersion = ''
    }

    let descriptors: Descriptor[] | undefined

    if (includeDescriptors) {
        const descriptorsResult = s.parseDescriptors()

        if (descriptorsResult instanceof Error) {
            return descriptorsResult
        }

        descriptors = descriptorsResult
    }

    return {
        scheme: scheme,
        package: {
            manager: manager,
            name: packageName,
            version: packageVersion,
        },
        descriptors: descriptors || [],
    }
}
