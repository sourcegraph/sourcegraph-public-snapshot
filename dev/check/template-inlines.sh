#!/bin/bash
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

set -euf -o pipefail
unset CDPATH

# Fails and prints matches if any HTML template files contain inline
# scripts or styles.

found=$(grep -EHnr '(<script|<style|style=)' app/templates | grep -v '<script src=' | grep -v '<script ignore-csp' | grep -v '<div ignore-csp' | grep -v '<style ignore-csp' || echo -n)

if [[ ! "$found" == "" ]]; then
    echo '!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!'
    echo '!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!'
    echo 'Found instances of inline script and style tags in HTML templates. These violate our CSP. Fix these!'
    echo '(See http://www.html5rocks.com/en/tutorials/security/content-security-policy/ for more info about CSP.)'
    echo '<script src="foo"> tags are OK, and <link rel="stylesheet" href=""> tags are OK. To make the former pass'
    echo 'this check script, put the src attribute immediately after "<script". (This script just uses a simple grep.)'
    echo '!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!'
    echo '!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!'
    echo "$found"
    exit 1
fi

exit 0
