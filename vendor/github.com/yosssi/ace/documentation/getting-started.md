# Getting Started

This document explains a basic usage of Ace by showing a simple example.

## 1. Create Ace templates

Create the following Ace templates.

base.ace

```ace
= doctype html
html lang=en
  head
    meta charset=utf-8
    title Ace example
    = css
      h1 { color: blue; }
  body
    h1 Base Template : {{.Msg}}
    #container.wrapper
      = yield main
      = yield sub
      = include inc .Msg
    = javascript
      alert('{{.Msg}}');
```

inner.ace

```ace
= content main
  h2 Inner Template - Main : {{.Msg}}

= content sub
  h3 Inner Template - Sub : {{.Msg}}
```

inc.ace

```ace
h4 Included Template : {{.}}
```

## 2. Create a web application

Create the following web application.

main.go

```go
package main

import (
	"net/http"

	"github.com/yosssi/ace"
)

func handler(w http.ResponseWriter, r *http.Request) {
	tpl, err := ace.Load("base", "inner", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tpl.Execute(w, map[string]string{"Msg": "Hello Ace"}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
```

## 3. Check the result.

Run the above web application and access [localhost:8080](http://localhost:8080). The following HTML will be rendered.

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Ace example</title>
    <style type="text/css">
      h1 { color: blue; }
    </style>
  </head>
  <body>
    <h1>Base Template : Hello Ace</h1>
    <div id="container" class="wrapper">
      <h2>Inner Template - Main : Hello Ace</h2>
      <h3>Inner Template - Sub : Hello Ace</h3>
      <h4>Included Template : Hello Ace</h4>
    </div>
    <script type="text/javascript">
      alert('Hello Ace');
    </script>
  </body>
</html>
```
