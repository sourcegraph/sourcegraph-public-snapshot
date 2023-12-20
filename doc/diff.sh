#!/usr/bin/env bash
set -eu

echo "Test"
sumFile="_docs.sum"
docsSum=$(find -L doc -name "docs.sum" -type f)
docFiles=$(find -L doc -name "*.md" -type f)
for f in ${docFiles}; do
    shasum ${f} >> ${sumFile}
done

sort -k 2 ${sumFile} > temp.txt
mv temp.txt ${sumFile}

sort -k 2 ${docsSum} > sorted_sums.sum

shasum ${sumFile}
shasum ${docsSum}

a="$(shasum ${sumFile} | cut -d ' ' -f 1)"
b="$(shasum  ${docsSum} | cut -d ' ' -f 1)"

# Compare files line by line using paste
line_number=0
diff_found=0
paste "${sumFile}" "sorted_sums.sum" | while IFS=$'\t' read -r line1 line2; do
    if [ "$line1" != "$line2" ]; then
        echo "Files differ: '$line1' vs '$line2'"
        diff_found=1
    fi
done
exit 1
