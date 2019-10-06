// TODO!(sqs): check this is comprehensive at https://docs.gradle.org/current/dsl/org.gradle.api.artifacts.dsl.DependencyHandler.html
export const parseDependencyNotation = (str: string): { group: string; name: string; version?: string } => {
    const parts = str.split(':', 3)
    if (parts.length === 2) {
        return { group: parts[0], name: parts[1] }
    }
    if (parts.length === 3) {
        return { group: parts[0], name: parts[1], version: parts[2] }
    }
    throw new Error(`invalid dependencyNotation ${JSON.stringify(str)}`)
}
