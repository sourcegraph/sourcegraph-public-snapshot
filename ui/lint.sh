#!/bin/bash
set -e

files=$(find ./web_modules -name '*.ts' -and -not -name '*.d.ts')
if [ ${files[@]} ]; then
	echo "Please only use .tsx or .d.ts files:"
	echo $files
	exit 1
fi

if grep -qr propTypes web_modules/; then
	echo "Please use type parameters instead of propTypes:"
	grep -rn propTypes web_modules/
	exit 1
fi

if grep -qr 'className={`' web_modules/; then
	echo "Please use the classNames helper instead of template strings:"
	grep -rn 'className={`' web_modules/
	exit 1
fi

# check for TypeScript errors before tslint
./node_modules/.bin/tsc
find ./web_modules -name '*.ts' -or -name '*.tsx' | xargs ./node_modules/.bin/tslint
