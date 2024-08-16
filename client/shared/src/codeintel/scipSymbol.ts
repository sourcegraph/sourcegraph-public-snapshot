import P from 'parsimmon'

export type Symbol = LocalSymbol | NonLocalSymbol

export interface LocalSymbol {
    kind: 'local'
    localID: string
}

export interface NonLocalSymbol {
    kind: 'nonlocal'
    scheme: string
    package: Package
    descriptors: Descriptor[]
}

export interface Package {
    manager: string
    name: string
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

export function parseSymbol(s: string): Symbol {
    return Parse.symbol.tryParse(s)
}

export function formatSymbol(s: Symbol): string {
    return Format.formatSymbol(s)
}

namespace Parse {
    const simpleIdentifier = P.regexp(/[A-Za-z0-9_\+\-\$]+/).desc('simple identifier')

    const escapedIdentifier = P.noneOf('`')
        .or(P.string('``'))
        .many()
        .wrap(P.string('`'), P.string('`'))
        .tie()
        .map(s => s.replaceAll('``', '`'))
        .desc('escaped identifier')

    const spaceTerminatedString = P.noneOf(' ')
        .or(P.string('  '))
        .many()
        .tie()
        .map(s => s.replaceAll('  ', ' '))
        .desc('space terminated string')

    const space = P.string(' ').desc('space')

    const localSymbol = P.string('local ')
        .then(simpleIdentifier)
        .map(
            (localID): LocalSymbol => ({
                kind: 'local',
                localID,
            })
        )
        .desc('local symbol')

    const packageParser = P.seq(spaceTerminatedString, space, spaceTerminatedString, space, spaceTerminatedString)
        .map(([manager, , packageName, , version]): Package => ({ manager, name: packageName, version }))
        .desc('package')

    const name = P.alt(escapedIdentifier, simpleIdentifier).desc('name')

    const parameterDescriptor = name
        .wrap(P.string('('), P.string(')'))
        .map((name): Descriptor => ({ kind: 'parameter', name }))
        .desc('parameter descriptor')

    const typeParameterDescriptor = name
        .wrap(P.string('['), P.string(']'))
        .map((name): Descriptor => ({ kind: 'typeParameter', name }))
        .desc('type parameter descriptor')

    const suffixToDescriptorKind: Record<string, Descriptor['kind']> = {
        '/': 'namespace',
        '#': 'type',
        '.': 'term',
        ':': 'meta',
        '!': 'macro',
    }

    const namedDescriptor: P.Parser<Descriptor> = name
        .chain(name =>
            P.alt(
                P.regexp(/[\/#\.:!]/).map(c => ({ kind: suffixToDescriptorKind[c], name })),
                simpleIdentifier
                    .times(0, 1)
                    .wrap(P.string('('), P.string(')'))
                    .skip(P.string('.'))
                    .map(disambiguator => ({ kind: 'method', name, disambiguator: disambiguator.at(0) }))
            )
        )
        .desc('named descriptor')

    const descriptor = P.alt(parameterDescriptor, typeParameterDescriptor, namedDescriptor).desc('descriptor')

    const nonLocalSymbol = P.seq(spaceTerminatedString, space, packageParser, space, descriptor.atLeast(1))
        .map(
            ([scheme, , parsedPackage, , descriptors]): NonLocalSymbol => ({
                kind: 'nonlocal',
                scheme,
                package: parsedPackage,
                descriptors,
            })
        )
        .desc('non-local symbol')

    export const symbol = P.alt(localSymbol, nonLocalSymbol).desc('symbol')
}

namespace Format {
    function esc(s: string): string {
        return s.replace(' ', '  ')
    }

    function formatNonLocalSymbol(s: NonLocalSymbol): string {
        return `${esc(s.scheme)} ${formatPackage(s.package)} ${s.descriptors.map(formatDescriptor).join('')}`
    }

    function formatPackage(p: Package): string {
        return `${esc(p.manager)} ${esc(p.name)} ${esc(p.version)}`
    }

    function formatDescriptor(d: Descriptor): string {
        switch (d.kind) {
            case 'namespace':
                return `${escapeName(d.name)}/`
            case 'type':
                return `${escapeName(d.name)}#`
            case 'term':
                return `${escapeName(d.name)}.`
            case 'meta':
                return `${escapeName(d.name)}:`
            case 'macro':
                return `${escapeName(d.name)}!`
            case 'method':
                return `${escapeName(d.name)}(${d.disambiguator ?? ''})`
            case 'typeParameter':
                return `[${escapeName(d.name)}]`
            case 'parameter':
                return `(${escapeName(d.name)})`
        }
    }

    function escapeName(s: string): string {
        if (/^[A-Za-z_\+\-\$]+$/.test(s)) {
            return s
        }
        return '`' + s.replace('`', '``') + '`'
    }

    export function formatSymbol(s: Symbol): string {
        if (s.kind === 'local') {
            return `local ${s.localID}`
        }
        return formatNonLocalSymbol(s)
    }
}
