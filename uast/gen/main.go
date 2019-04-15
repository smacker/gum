// script to generate fixtures
package main

import (
	"fmt"

	bblfsh "gopkg.in/bblfsh/client-go.v3"
	uastyml "gopkg.in/bblfsh/sdk.v2/uast/yaml"
)

var srcContent = `public class Test {
  public String foo(int i) {
	if (i == 0) return "Foo!";
  }
}
`

var dstContent = `public class Test {
  public String foo(int i) {
    if (i == 0) return "Bar";
    else if (i == -1) return "Foo!";
  }
}
`

func main() {
	client, err := bblfsh.NewClient("0.0.0.0:9432")
	if err != nil {
		panic(err)
	}

	res, _, err := client.
		NewParseRequest().
		Mode(bblfsh.Annotated).
		Language("java").
		Content(srcContent).
		UAST()
	if err != nil {
		panic(err)
	}

	b, err := uastyml.Marshal(res)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
