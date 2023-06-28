$target = $args[0]

echo "Signing: ${target} | ${env:STORE_TYPE}/${env:STORE_KEY}/${env:STORE_ALIAS}/${env:STORE_CERT}"

if ($null -eq $env:STORE_TYPE) {
    throw "Missing STORE_TYPE"
}
if ($null -eq $env:STORE_PASS) {
    throw "Missing STORE_PASS"
}
if ($null -eq $env:STORE_ALIAS) {
    throw "Missing STORE_ALIAS"
}

if ($null -eq $env:STORE_CERT) {
    jsign --storetype "${env:STORE_TYPE}" --storepass "${env:STORE_PASS}" `
        --keystore "${env:STORE_KEY}" `
        --alias "${env:STORE_ALIAS}" `
        --alg SHA-256 `
        --tsaurl http://timestamp.digicert.com `
        "${target}"
}
else {
    jsign --storetype "${env:STORE_TYPE}" --storepass "${env:STORE_PASS}" `
        --keystore "${env:STORE_KEY}" `
        --alias "${env:STORE_ALIAS}" `
        --certfile ${env:STORE_CERT} `
        --alg SHA-256 `
        --tsaurl http://timestamp.digicert.com `
        "${target}"
}

if (!$?) {
    throw "Failed to sign '${target}'"
}
