package merger

import (
	"bytes"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	a := struct {
		One   string `json:"one,omitempty"`
		Two   int
		Three bool
	}{}
	b := struct {
		Two int
	}{2}
	c := map[string]string{"one": "c"}
	err := Merge(&a, b)
	if err != nil {
		t.Fatal(err)
	}
	err = Merge(&a, c)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "c", a.One)
	assert.Equal(t, 2, a.Two)
}

func TestComplex(t *testing.T) {
	type B2 struct {
		Name string
	}
	type A2 struct {
		Person B2
		Cars   int
	}
	type C2 struct {
		Person B2
		Cars   int
	}
	a := &A2{}
	a.Person.Name = "John"
	a.Cars = 2
	b := C2{}
	b.Person.Name = "Doe"
	b.Cars = 10
	err := Merge(a, b)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "Doe", a.Person.Name)
	assert.Equal(t, 10, a.Cars)
}

func TestMap(t *testing.T) {
	a := make(map[string]interface{})
	a["t"] = 1
	b := struct {
		A          string `json:"t"`
		B          float64
		unexported int
	}{
		"test",
		10.2,
		1,
	}
	err := merge(&a, b, true)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "test", a["t"])
	assert.Equal(t, float64(10.2), a["B"])
}

func TestTime(t *testing.T) {
	bff := new(bytes.Buffer)
	log.SetOutput(bff)
	a := make(map[string]interface{})
	a["t"] = "2015-01-01T18:39:18.379414425-03:00"
	b := struct {
		T time.Time `json:"t"`
	}{}
	err := merge(&b, a, true)
	assert.NoError(t, err)
	t.Log(bff.String())
	assert.Equal(t, 2015, b.T.Year())

}
