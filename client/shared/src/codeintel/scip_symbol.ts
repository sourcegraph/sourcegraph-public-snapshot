import {
    buildLexer,
    kmid,
    seq,
    tok,
    expectEOF,
    apply,
    kright,
    str,
    nil,
    alt_sc,
    list_n,
    list_sc,
    opt,
    expectSingleResult,
} from 'typescript-parsec'

export type Symbol = LocalSymbol | NonLocalSymbol

export type LocalSymbol = {
    local_id: string
}

export type NonLocalSymbol = {
    scheme: string
    package: Package
    descriptors: Descriptor[]
}

export type Package = {
    manager: string
    package_name: string
    version: string
}

export type Descriptor =
    | { kind: 'namespace'; name: string }
    | { kind: 'type'; name: string }
    | { kind: 'term'; name: string }
    | { kind: 'meta'; name: string }
    | { kind: 'macro'; name: string }
    | { kind: 'method'; name: string; disambiguator?: string }
    | { kind: 'typeParameter'; name: string }
    | { kind: 'parameter'; name: string }

export function parseSymbolName(symbolName: string): Symbol {
    return expectSingleResult(expectEOF(symbolParser.parse(tokenizer.parse(symbolName))))
}

enum TokenKind {
    KeywordLocal,
    Space,
    SimpleIdentifier,
    SpacelessString,
    EscapedIdentifier,
    OpenParen,
    ClosedParen,
    OpenBracket,
    ClosedBracket,
    Dot,
    SimpleDescriptorSuffix,
}

const tokenizer = buildLexer([
    [true, /^local/g, TokenKind.KeywordLocal],
    [true, /^ /g, TokenKind.Space],
    [true, /^`[^`]*`+/g, TokenKind.EscapedIdentifier],
    [true, /^[A-Za-z0-9_\-\+\$]+/g, TokenKind.SimpleIdentifier],
    [true, /^[^ ]+/g, TokenKind.SpacelessString],
    [true, /^\(/g, TokenKind.OpenParen],
    [true, /^\)/g, TokenKind.ClosedParen],
    [true, /^\[/g, TokenKind.OpenBracket],
    [true, /^\]/g, TokenKind.ClosedBracket],
    [true, /^\./g, TokenKind.Dot],
    [true, /^[\/#\.:!]/g, TokenKind.SimpleDescriptorSuffix],
])

const localSymbolParser = apply(
    kright(str('local '), tok(TokenKind.SimpleIdentifier)),
    (id): LocalSymbol => ({ local_id: id.text })
)

const packageParser = apply(
    list_n(tok(TokenKind.SpacelessString), tok(TokenKind.Space), 3),
    ([manager, name, version]): Package => ({ manager: manager.text, package_name: name.text, version: version.text })
)

const nameParser = apply(alt_sc(tok(TokenKind.EscapedIdentifier), tok(TokenKind.SimpleIdentifier)), t => t.text)

const parameterDescriptorParser = apply(
    kmid(str('('), nameParser, str(')')),
    (name): Descriptor => ({ kind: 'parameter', name })
)

const typeParameterDescriptorParser = apply(
    kmid(str('['), nameParser, str(']')),
    (name): Descriptor => ({ kind: 'parameter', name })
)

const suffixToDescriptorKind: Record<string, Descriptor['kind']> = {
    '/': 'namespace',
    '#': 'type',
    '.': 'term',
    ':': 'meta',
    '!': 'macro',
}

const namedDescriptorParser = apply(
    seq(
        nameParser,
        alt_sc(
            tok(TokenKind.SimpleDescriptorSuffix),
            seq(str('('), opt(tok(TokenKind.SimpleIdentifier)), str(')'), str('.'))
        )
    ),
    ([name, suffix]): Descriptor => {
        if (suffix instanceof Array) {
            const [, disambiguator] = suffix
            return { kind: 'method', name, disambiguator: disambiguator?.text }
        }
        return { kind: suffixToDescriptorKind[suffix.text], name }
    }
)

const descriptorParser = alt_sc(parameterDescriptorParser, typeParameterDescriptorParser, namedDescriptorParser)

const nonLocalSymbolParser = apply(
    seq(tok(TokenKind.SpacelessString), str(' '), packageParser, str(' '), list_sc(descriptorParser, nil())),
    ([scheme, , pack, , descriptors]): NonLocalSymbol => ({ scheme: scheme.text, package: pack, descriptors })
)

const symbolParser = alt_sc(localSymbolParser, nonLocalSymbolParser)
