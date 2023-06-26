$target = $args[0]

echo "Signing: ${target} | ${env:STORE_TYPE}/${env:STORE_KEY}/${env:STORE_ALIAS}/${env:STORE_CERT}"

if ($env:STORE_TYPE -eq $null) {
    throw "Missing STORE_TYPE"
}
if ($env:STORE_PASS -eq $null) {
    throw "Missing STORE_PASS"
}
if ($env:STORE_ALIAS -eq $null) {
    throw "Missing STORE_ALIAS"
}
if (! $env:STORE_CERT -eq $null) {
    $cert="--certfile ${env:STORE_CERT}"
    throw "Missing STORE_CERT"
}

jsign --storetype "${env:STORE_TYPE}" --storepass "${env:STORE_PASS}" `
    --keystore "${env:STORE_KEY}" `
    --alias "${env:STORE_ALIAS}" `
    $cert `
    --alg SHA-256 `
    --tsaurl http://timestamp.digicert.com `
    "${target}"
if (!$?) {
    throw "Failed to sign '${target}'"
}
