+++
title = "Content security policy"
+++

First, read
http://www.html5rocks.com/en/tutorials/security/content-security-policy/.

CSPs are a way we can protect against XSS. Our CSP policy forbids
inline scripts, styles, eval, and so on. The
`dev/check-for-template-inlines` script checks for these, but it won't
find everything. So, don't use inline scripts or styles!
