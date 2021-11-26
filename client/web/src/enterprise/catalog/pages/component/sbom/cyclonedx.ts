export interface BOM {
    metadata?: Metadata
    components?: Component[]
    dependencies?: Dependency[]
}

// eslint-disable-next-line unicorn/prevent-abbreviations
export type BOMRef = string

export interface Metadata {
    timestamp?: string
    component?: Component
    licenses?: LicenseChoice[]
}

export interface Component {
    type: 'application' | 'library'
    name: string
    version: string
    description?: string
    'bom-ref'?: BOMRef
    purl?: string
    externalReferences: ExternalReference[]
    scope?: 'required' | 'optional' | 'excluded'
}

export interface ExternalReference {
    url: string
    type: string
}

export interface Evidence {
    licenses?: LicenseChoice[]
}

export type LicenseChoice = { license: License } | { expression: string }

export type License = { id: string } | { name: string }

export interface Dependency {
    ref: BOMRef
    dependsOn?: BOMRef[]
}
