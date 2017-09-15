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
	//
	// merge two maps
	{
		mapa := map[string]interface{}{
			"a": 10,
			"b": "bee",
			"c": false,
			"d": 0,
			"e": 0,
		}
		mapb := map[string]int{
			"d": -10,
			"e": 100,
			"f": 2,
		}
		assert.NoError(t, merge(&mapa, &mapb, true))
		assert.Equal(t, -10, mapa["d"])
		assert.Equal(t, 100, mapa["e"])
		assert.Equal(t, 2, mapa["f"])
	}
}

func TestTime(t *testing.T) {
	bff := new(bytes.Buffer)
	log.SetOutput(bff)
	a := make(map[string]interface{})
	a["t"] = "2015-01-01T18:39:18.379414425-03:00"
	a["r"] = int64(100)
	a["r2"] = float64(100)
	b := struct {
		T  time.Time `json:"t"`
		R  int64     `json:"r"`
		R2 int       `json:"r2"`
	}{}
	err := merge(&b, a, true)
	assert.NoError(t, err)
	t.Log(bff.String())
	assert.Equal(t, 2015, b.T.Year())
	assert.Equal(t, int64(100), b.R)
	assert.Equal(t, 100, b.R2)

}

func TestOverwrite(t *testing.T) {
	a := struct {
		Firstname string
		Lastname  string `json:"lastname"`
		Age       int
	}{}
	a.Firstname = "Gabriel"
	a.Age = 30
	b := struct {
		Age   int
		Hands int
	}{25, 2}
	c := map[string]interface{}{
		"Firstname": "John",
		"lastname":  "Doe",
	}
	err := MergeOverwrite(&a, b)
	assert.NoError(t, err)
	err = MergeOverwrite(&a, c)
	assert.NoError(t, err)
	assert.Equal(t, c["Firstname"], a.Firstname)
	assert.Equal(t, c["lastname"], a.Lastname)
	assert.Equal(t, b.Age, a.Age)
}

func TestFail(t *testing.T) {
	assert.EqualError(t, merge(nil, nil, false), "dst cannot be nil")
	m := 1
	assert.EqualError(t, merge(&m, nil, false), "src cannot be nil")
	assert.EqualError(t, merge(&m, &m, false), "invalid dst kind int")
	n := map[string]interface{}{}
	assert.EqualError(t, merge(n, &m, false), "dst needs to be a pointer")
	assert.EqualError(t, merge(m, &m, false), "invalid destination kind int")
	assert.EqualError(t, merge(&n, m, false), "invalid source kind int")
}

func TestNumericConversion(t *testing.T) {
	{
		a := struct {
			A int
			B int
		}{}
		b := struct {
			A int16
		}{}
		b.A = 10
		assert.NoError(t, MergeOverwrite(&a, &b))
		assert.NotEqual(t, 10, a.A) // should not force conversion on struct to struct
	}
	{
		a := struct {
			A int
			B int
		}{}
		b := map[string]interface{}{
			"A": int16(10),
			"B": uint8(3),
		}
		assert.NoError(t, MergeOverwrite(&a, &b))
		assert.Equal(t, 10, a.A) // should force conversion on map to struct
		assert.Equal(t, 3, a.B)  // should force conversion on map to struct
	}
}
