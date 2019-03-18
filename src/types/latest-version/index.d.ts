declare module 'latest-version' {
    declare function latestVersion(packageName: string): Promise<string>
    export = latestVersion
}
