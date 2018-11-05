import * as sourcegraph from 'sourcegraph'

export class URI implements sourcegraph.URI {
    public static parse(uri: string): sourcegraph.URI {
        return new URI(uri)
    }

    public static file(path: string): sourcegraph.URI {
        return new URI(`file://${path}`)
    }

    public static isURI(value: any): value is sourcegraph.URI {
        return value instanceof URI || typeof value === 'string'
    }

    constructor(private value: string) {}

    public toString(): string {
        return this.value
    }

    public toJSON(): any {
        return this.value
    }
}
