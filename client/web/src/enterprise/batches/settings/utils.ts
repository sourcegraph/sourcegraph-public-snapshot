/** Parses and returns the name from a batch spec YAML file, if it has one */
export const parseNameFromSpec = (spec: string): string =>
    // First capture group will be the name
    spec.match(/name:\s*(\S*)/)?.[1] || '-'
