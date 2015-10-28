package syntaxhighlight_test

import (
	"fmt"
	"os"

	"github.com/sourcegraph/syntaxhighlight"
)

func Example() {
	src := []byte(`
/* hello, world! */
var a = 3;

// b is a cool function
function b() {
  return 7;
}`)

	highlighted, err := syntaxhighlight.AsHTML(src)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(string(highlighted))

	// Output:
	// <span class="com">/* hello, world! */</span>
	// <span class="kwd">var</span> <span class="pln">a</span> <span class="pun">=</span> <span class="dec">3</span><span class="pun">;</span>
	//
	// <span class="com">// b is a cool function</span>
	// <span class="kwd">function</span> <span class="pln">b</span><span class="pun">(</span><span class="pun">)</span> <span class="pun">{</span>
	//   <span class="kwd">return</span> <span class="dec">7</span><span class="pun">;</span>
	// <span class="pun">}</span>
}
