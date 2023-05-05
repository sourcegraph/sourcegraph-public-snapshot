export interface ISymbol {
    scheme: string
    package?: Package
    descriptors: Descriptor[]
}

export interface SCIPOccurrence {
    symbol: string
}

export interface SCIPDocument {
    relative_path: string
    occurrences: SCIPOccurrence[]
    symbols: { symbol: string }[]
}

export interface SCIPSymbol {
    scheme: string
    package?: Package
    descriptors?: Descriptor[]
}

interface Package {
    manager: string
    name: string
    version: string
}

interface Descriptor {
    name: string
    disambiguator?: string
    suffix: DescriptorSuffix
}

export enum DescriptorSuffix {
    UnspecifiedSuffix = 0,

    // Unit of code abstraction and/or namespacing.
    //
    // NOTE: This corresponds to a package in Go and JVM languages.
    Namespace = 1,

    // Use Namespace instead.
    //
    // Deprecated: Do not use.
    Package = 1,

    Type = 2,
    Term = 3,
    Method = 4,
    TypeParameter = 5,
    Parameter = 6,

    // Can be used for any purpose.
    Meta = 7,
    Local = 8,
    Macro = 9,
}

export function parseSymbol(symbol: string): SCIPSymbol {
    return parsePartialSymbol(symbol, true)
}

function parsePartialSymbol(symbol: string, includeDescriptors: boolean): SCIPSymbol {
    const parser = new SymbolParser(symbol)
    const scheme = parser.acceptSpaceEscapedIdentifier('scheme')
    if (scheme === 'local') {
        return newLocalSymbol(symbol.slice(parser.index))
    }
    let manager = parser.acceptSpaceEscapedIdentifier('package manager')
    if (manager === '.') {
        manager = ''
    }
    let packageName = parser.acceptSpaceEscapedIdentifier('package name')
    if (packageName === '.') {
        packageName = ''
    }
    let packageVersion = parser.acceptSpaceEscapedIdentifier('package version')
    if (packageVersion === '.') {
        packageVersion = ''
    }
    return {
        scheme,
        package: { manager, name: packageName, version: packageVersion },
        descriptors: includeDescriptors ? parser.parseDescriptors() : undefined,
    }
}

class SymbolParser {
    public index = 0

    constructor(private symbol: string) {}

    private current(): string {
        return this.symbol[this.index]
    }

    public acceptSpaceEscapedIdentifier(what: string): string {
        return this.acceptEscapedIdentifier(what, ' ')
    }

    private acceptEscapedIdentifier(what: string, escapeCharacter: string): string {
        let identifier = ''
        while (this.index < this.symbol.length) {
            const char = this.current()
            if (char === escapeCharacter) {
                this.index++

                if (this.index >= this.symbol.length) {
                    break
                }

                if (this.current() === escapeCharacter) {
                    // Escaped space character.
                    identifier += char
                } else {
                    return identifier
                }
            } else {
                identifier += char
            }
            this.index++
        }
        throw new Error(`Reached end of symbol while parsing ${what}, expected a '${escapeCharacter}' character`)
    }

    public parseDescriptors(): Descriptor[] {
        const descriptors = []
        while (this.index < this.symbol.length) {
            const descriptor = this.parseDescriptor()
            if (descriptor) {
                descriptors.push(descriptor)
            }
        }
        return descriptors
    }

    private peekNext(): string | null {
        if (this.index + 1 < this.symbol.length) {
            return this.symbol[this.index]
        }
        return null
    }

    private parseDescriptor(): Descriptor | null {
        // debugger
        switch (this.peekNext()) {
            // case null:
            //     this.index++
            //     return null

            case '(': {
                this.index++
                const name = this.acceptIdentifier('parameter name')
                // TODO: handle error - return null?
                this.acceptCharacter(')', 'closing parameter name')
                return { name, suffix: DescriptorSuffix.Parameter }
            }

            case '[': {
                this.index++
                const name = this.acceptIdentifier('type parameter name')
                // TODO: handle error - return null?
                this.acceptCharacter(']', 'closing type parameter name')
                return { name, suffix: DescriptorSuffix.TypeParameter }
            }

            default: {
                const name = this.acceptIdentifier('descriptor name')

                // console.log(name)

                // TODO: handle error - return null?
                const suffix = this.current()
                this.index++

                // console.log({ name, slice: this.symbol.slice(this.index), suffix })

                // if (!name) {
                // }
                switch (suffix) {
                    case '(': {
                        let disambiguator: string | undefined = undefined
                        if (this.peekNext() !== ')') {
                            disambiguator = this.acceptIdentifier('method disambiguator')
                            // TODO: handle error - return null?
                        }
                        this.acceptCharacter(')', 'closing method')
                        // TODO: handle error - return null?

                        this.acceptCharacter('.', 'closing method')

                        return { name, disambiguator, suffix: DescriptorSuffix.Method }
                    }
                    case '/':
                        return { name, suffix: DescriptorSuffix.Namespace }
                    case '.':
                        return { name, suffix: DescriptorSuffix.Term }
                    case '#':
                        return { name, suffix: DescriptorSuffix.Type }
                    case ':':
                        return { name, suffix: DescriptorSuffix.Meta }
                    case '!':
                        return { name, suffix: DescriptorSuffix.Macro }
                    default:
                }
            }
        }
        return null
    }

    private acceptCharacter(char: string, what: string): null {
        if (this.current() === char) {
            this.index++
            return null
        }
        throw new Error(`Expected '${char}', obtained '${this.current()}', while parsing ${what}`)
    }

    private acceptIdentifier(what: string): string {
        if (this.current() == '`') {
            this.index++
            return this.acceptBacktickEscapedIdentifier(what)
        }
        const start = this.index
        while (this.index < this.symbol.length && isIdentifierCharacter(this.current())) {
            this.index++
        }
        if (start === this.index) {
            // console.log(this.symbol, what)
            return ''
            //  throw new Error('empty identifier')
        }
        return this.symbol.slice(start, this.index)
    }

    private acceptBacktickEscapedIdentifier(what: string): string {
        return this.acceptEscapedIdentifier(what, '`')
    }
}

function isIdentifierCharacter(char: string) {
    return /^[a-z0-9-+$_]$/i.test(char)
}

function newLocalSymbol(id: string): SCIPSymbol {
    return {
        scheme: 'local',
        descriptors: [{ name: id, suffix: DescriptorSuffix.Local }],
    }
}
