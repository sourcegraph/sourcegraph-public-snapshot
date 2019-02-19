import * as sourcegraph from 'sourcegraph'

export class URI implements sourcegraph.URI {
    public static parse(uri: string): URI {
        return new URI(uri)
    }

    public static file(path: string): URI {
        return new URI(`file://${path}`)
    }

    public static isURI(value: any): value is URI {
        return value instanceof URI || typeof value === 'string' // TODO this is blatandly wrong, strings are not URI objects!
    }

    constructor(private value: string) {}

    public toString(): string {
        return this.value
    }

    public toJSON(): any {
        return this.value
    }
}
