import * as React from 'react'

// Production extension Url
const extensionUrl = 'chrome-extension://dgjhfomjieaadpoljlnidmbgkdffpack'

// Dev extension url
// const extensionUrl = 'chrome-extension://bmfbcejdknlknpncfpeloejonjoledha'

export const LinkExtension: React.SFC<{}> = () => (
    <embed className="link-extension" src={`${extensionUrl}/link.html?sourceurl=${window.location.origin}`} />
)
