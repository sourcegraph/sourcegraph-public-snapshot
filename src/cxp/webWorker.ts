/**
 * Returns a data: URL to use as the entry point of a new Web Worker. The Web Worker will immediately load the
 * script at the given URL.
 */
export function importScriptsBlobURL(id: string, scriptURL: string): string {
    const b = new Blob(
        [`try{importScripts('${new URL(scriptURL).toString()}')}catch(err){console.error(err);throw err}`],
        {
            type: 'application/javascript',
        }
    )
    return window.URL.createObjectURL(b) + `#${id}`
}
