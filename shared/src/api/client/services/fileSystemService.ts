/**
 * The file system service manages file systems.
 */
export interface FileSystemService {
    /**
     * Read the contents of the resource at the URI.
     */
    readFile(uri: URL): Promise<string>

    setProvider(provider: (uri: URL) => Promise<string>): void
}

/**
 * Creates a new instance of {@link FileSystemService}.
 */
export function createFileSystemService(): FileSystemService {
    let provider: ((uri: URL) => Promise<string>) | undefined
    return {
        readFile: uri => {
            if (!provider) {
                throw new Error('no file system provider registered')
            }
            return provider(uri)
        },
        setProvider: p => {
            provider = p
        },
    }
}
