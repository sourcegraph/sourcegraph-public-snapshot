import { ajax } from 'rxjs/ajax'

export async function createBlobURLForBundle(bundleURL: string): Promise<string> {
    const req = await ajax({
        url: bundleURL,
        crossDomain: true,
        responseType: 'blob',
    }).toPromise()
    return window.URL.createObjectURL(req.response)
}
