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
	err := merge(&a, b, true, "json", nil)
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
		assert.NoError(t, merge(&mapa, &mapb, true, "json", nil))
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
	err := merge(&b, a, true, "json", nil)
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
	assert.EqualError(t, merge(nil, nil, false, "json", nil), "dst cannot be nil")
	m := 1
	assert.EqualError(t, merge(&m, nil, false, "json", nil), "src cannot be nil")
	assert.EqualError(t, merge(&m, &m, false, "json", nil), "invalid dst kind int")
	n := map[string]interface{}{}
	assert.EqualError(t, merge(n, &m, false, "json", nil), "dst needs to be a pointer")
	assert.EqualError(t, merge(m, &m, false, "json", nil), "invalid destination kind int")
	assert.EqualError(t, merge(&n, m, false, "json", nil), "invalid source kind int")
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
		assert.Equal(t, 10, a.A) // should force conversion on struct to struct
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

type TestStr string

func TestTag(t *testing.T) {
	a := struct {
		ID          int     `db:"id"`
		Name        TestStr `db:"name"`
		Description string  `db:"desc"`
	}{
		10,
		"Books",
		"This is a category",
	}
	b := map[string]interface{}{
		"score": 10,
	}
	assert.NoError(t, MergeWithTag(&b, &a, "db"))
	assert.Equal(t, int(10), b["id"])
	assert.Equal(t, TestStr("Books"), b["name"])
	assert.NotEqual(t, "Books", b["name"]) // TestStr is not equal to string
	assert.Equal(t, "This is a category", b["desc"])
}

func TestOmitEmpty(t *testing.T) {
	a := struct {
		Alpha string `json:"alpha,omitempty"`
		Beta  string `json:"beta"`
	}{
		"",
		"2",
	}
	b := map[string]interface{}{
		"charlie": "3",
	}
	assert.NoError(t, Merge(&b, &a))
	if _, ok := b["alpha"]; ok {
		t.Error("beta is present")
	}
}

func TestCasing(t *testing.T) {
	a := struct {
		ProtonID string `json:"proton_id"`
		FinalUrl string `json:"final_url"`
	}{}
	b := struct {
		ProtonId string `json:"proton_id"`
		FinalURL string `json:"final_url"`
	}{}
	a.FinalUrl = "https://www.google.com/?q=golang"
	b.ProtonId = "123456"
	assert.NoError(t, MergeOverwriteWithTag(&b, &a, "json"))
	assert.Equal(t, a.ProtonID, b.ProtonId)
	assert.Equal(t, a.FinalUrl, b.FinalURL)
}
