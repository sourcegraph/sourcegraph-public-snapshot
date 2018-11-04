declare module 'plist' {
    // parse takes in the contents of a .plist file and returns the values in an array
    export function parse(data: string): any[]

    // build takes data and turns it into a .plist
    export function build(data: any): string
}
