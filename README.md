# Merger
```Go
import(
	"fmt"

	"github.com/gabstv/merger"
)

type A struct {
	Foo string
	Bar string `json:"bar"`
	Baz int `json:"baz"`
}

type B struct {
	Bar string
}

func main(){
	a := A{}
	a.Foo = "foo"
	b := B{}
	b.Bar = "bar"
	merger.Merge(&a, b)
	merger.Merge(&a, map[string]int{
		"baz": 1,
	})
	fmt.Println(a.Foo)
	fmt.Println(a.Bar)
	fmt.Println(a.Baz)
}
```