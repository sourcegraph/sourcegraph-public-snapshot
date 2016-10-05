package auth

import (
	"testing"
)

func TestToken(t *testing.T) {
	tok, err := NewAccessToken(nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ParseAndVerify(tok); err != nil {
		t.Fatal(err)
	}

	defaultKey := ActiveIDKey
	defer func() {
		ActiveIDKey = defaultKey
	}()
	ActiveIDKey, err = NewIDKey([]byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAp0CJyHipPd63h/+kI1aLcg6EzERUuACLDmpj7JfZVLEz2rkp
gGJO+iA1mDLrv485rsYVhoGgsqdGOc8PNU4D9IcHJ4VXU6ZNTlx/uR7HUpwkGITH
C4fIsTJvHEgc2w3pNfWKOVimubR2V6E6qh+pRnqGzVzM/Rp5guuKm3DJyeBia3mv
oM/IBURojLXbFcB5SAGyC+Fv9Mkz2ycBjk+31FiEdnsJFC6I/QmeK0z66qGIfojq
rNql5+Zd/kUNQtjZlF9uFcu/NIiQgI6aKkndH7mjOunKniaMxqDEAMn09uZm7rBe
Ucr61jT0CHR8rvY8pYEl3HeMomgzDDPbTOVjCQIDAQABAoIBAAmfAs4PctzmRPSD
1jNaNSdYgnclryHulhE8OYdQrOXcU7lPUX3bKePlmm+o7jrUyGKvbmmQZ2gfi0Ck
EqHkXQHiCp1RZFahiGzrkUVa6ehspv7qFHErXHYlCpM76r0HLdU2zL7DxMOGCBC+
a5uBusEdJ0gFAJ3Guhq35f9PG6yLLg2aEicSnFISLmA9gBy05onyawnOYWK8Jb6Q
b5YIDFEg6W/f/24EPmc5Vto+Ux6Cxag7tY82GKYJG02KKk25wMlZBGUhxEpMmaG7
T63OGUlZ+jiXx5f4yyuyLsfc79zpKxMv0Kf7w194zPpxQchXfUT0/JkdxgG0wMxb
NUQ1w2kCgYEAyYn/nnW5oDJyEyEh3mHZADULXWOMRp7KZD9uqqMdzxOR0i7K2qCV
jmm1ok2clpfbYaUMZeIC1jHhD3Op32WCHx+OkiU8Q1o1i/4c6xCp0RsFyivkchfR
ChvMqvNsFoo8jvsDIYoyLjTC+6laGHfL3J0b6aFOEtuxg4W38wAjU7sCgYEA1HKl
8zZxUxmuCtUcLfzDznC9xGnDoQsyQAKD6DPgq/dsexaQUERbs0/ZNRHlQ87pLdfm
U/ly6i9QlwUBlvz6eq/n9tHbTfHlBFnosVVCrAnpywgYNekKd1yle6V98DnNBOFq
+SkUSx7rwqR8S2PCpLjLQhBbYI0UOUvQZ//0vgsCgYEAlr40RtaxOARjVLGUfpxb
Tg9e58Q8uNmuclsLsG//LNLrX/WF3w77riCdLb+1XuJIwflMk6wACSwXtZICvkhT
kmntHpzhPVNs97/i62N0USZQJ067OSddQJ1YcYlPEHDnKN7REbYnIG5wZQHflKuN
/P46UX5IQky2srRCyWwSAF8CgYBPgtk5PZcMUwAgbcIuM/vUt71OVYcyLs6PxmE3
9rKPqfqf1sIMSIlJgwj4I8p6pmX/El7R7vpjS3IOE4GU0PmuEUfvyHsboPzltACy
3gYl/U/S/SSSiLWyFqqYrEeGMRvaR8ORnR5LPzddkdIzJRMkM0VfZF/Osv5us0E8
qz8eIQKBgGmrEOZVTbLN4gLBkXJ6Ji+GAcguBuouJvxpzD9wAMTKwxLAIAIIuWf/
4GlSJJvuvoliMzHSbz1KpF2CxABSz4gRsuCKjx4QDtH0fiUoqHTUu1P5QweXKHPC
13GbHbtXDNKL66FkqMbJGI+owbdnfZS1X7TKyAhuK2JMYmyctGqN
-----END RSA PRIVATE KEY-----`))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := ParseAndVerify(tok); err == nil {
		t.Fatal("error expected")
	}
}
