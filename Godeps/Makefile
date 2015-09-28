update-to-new-go-subrepo-paths:
	rm -rf _workspace/src/code.google.com/p/go.*
	find _workspace/src -name '*.go' -type f | xargs sed -i 's/code.google.com\/p\/go\./golang.org\/x\//g'
