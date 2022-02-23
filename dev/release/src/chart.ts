import { readFileSync } from 'fs'

import { load as loadYAML } from 'js-yaml'

export interface Metadata {
    apiVersion: string
    name: string
    description: string
    type: string
    version: string
    appVersion: string
}

export function parseChartMetadata(chartYamlPath: string): Metadata {
    const chartYamlContents = readFileSync(chartYamlPath, 'utf8').toString()
    return loadYAML(chartYamlContents) as Metadata
}
