package flags

import (
	"testing"
	"time"
)

func expectConvert(t *testing.T, o *Option, expected string) {
	s, err := convertToString(o.value, o.tag)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	assertString(t, s, expected)
}

func TestConvertToString(t *testing.T) {
	d, _ := time.ParseDuration("1h2m4s")

	var opts = struct {
		String string `long:"string"`

		Int   int   `long:"int"`
		Int8  int8  `long:"int8"`
		Int16 int16 `long:"int16"`
		Int32 int32 `long:"int32"`
		Int64 int64 `long:"int64"`

		Uint   uint   `long:"uint"`
		Uint8  uint8  `long:"uint8"`
		Uint16 uint16 `long:"uint16"`
		Uint32 uint32 `long:"uint32"`
		Uint64 uint64 `long:"uint64"`

		Float32 float32 `long:"float32"`
		Float64 float64 `long:"float64"`

		Duration time.Duration `long:"duration"`

		Bool bool `long:"bool"`

		IntSlice    []int           `long:"int-slice"`
		IntFloatMap map[int]float64 `long:"int-float-map"`

		PtrBool   *bool       `long:"ptr-bool"`
		Interface interface{} `long:"interface"`

		Int32Base  int32  `long:"int32-base" base:"16"`
		Uint32Base uint32 `long:"uint32-base" base:"16"`
	}{
		"string",

		-2,
		-1,
		0,
		1,
		2,

		1,
		2,
		3,
		4,
		5,

		1.2,
		-3.4,

		d,
		true,

		[]int{-3, 4, -2},
		map[int]float64{-2: 4.5},

		new(bool),
		float32(5.2),

		-5823,
		4232,
	}

	p := NewNamedParser("test", Default)
	grp, _ := p.AddGroup("test group", "", &opts)

	expects := []string{
		"string",
		"-2",
		"-1",
		"0",
		"1",
		"2",

		"1",
		"2",
		"3",
		"4",
		"5",

		"1.2",
		"-3.4",

		"1h2m4s",
		"true",

		"[-3, 4, -2]",
		"{-2:4.5}",

		"false",
		"5.2",

		"-16bf",
		"1088",
	}

	for i, v := range grp.Options() {
		expectConvert(t, v, expects[i])
	}
}

func TestConvertToStringInvalidIntBase(t *testing.T) {
	var opts = struct {
		Int int `long:"int" base:"no"`
	}{
		2,
	}

	p := NewNamedParser("test", Default)
	grp, _ := p.AddGroup("test group", "", &opts)
	o := grp.Options()[0]

	_, err := convertToString(o.value, o.tag)

	if err != nil {
		err = newErrorf(ErrMarshal, "%v", err)
	}

	assertError(t, err, ErrMarshal, "strconv.ParseInt: parsing \"no\": invalid syntax")
}

func TestConvertToStringInvalidUintBase(t *testing.T) {
	var opts = struct {
		Uint uint `long:"uint" base:"no"`
	}{
		2,
	}

	p := NewNamedParser("test", Default)
	grp, _ := p.AddGroup("test group", "", &opts)
	o := grp.Options()[0]

	_, err := convertToString(o.value, o.tag)

	if err != nil {
		err = newErrorf(ErrMarshal, "%v", err)
	}

	assertError(t, err, ErrMarshal, "strconv.ParseInt: parsing \"no\": invalid syntax")
}

func TestConvertToMapWithDelimiter(t *testing.T) {
	var opts = struct {
		StringStringMap map[string]string `long:"string-string-map" key-value-delimiter:"="`
	}{}

	p := NewNamedParser("test", Default)
	grp, _ := p.AddGroup("test group", "", &opts)
	o := grp.Options()[0]

	err := convert("key=value", o.value, o.tag)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	assertString(t, opts.StringStringMap["key"], "value")
}

type testEnum int

const (
	one testEnum = iota
	two
	three
)

func (t *testEnum) UnmarshalText(text []byte) error {
	switch string(text) {
	case "one":
		*t = one
	case "two":
		*t = two
	case "three":
		*t = three
	default:
		return newErrorf(ErrMarshal, "invalid value %q", text)
	}
	return nil
}

func (t testEnum) MarshalText() ([]byte, error) {
	switch t {
	case one:
		return []byte("one"), nil
	case two:
		return []byte("two"), nil
	case three:
		return []byte("three"), nil
	default:
		return nil, newErrorf(ErrMarshal, "invalid value %q", t)
	}
}

func TestConvertUsesUnmarshalText(t *testing.T) {
	var opt = struct {
		Enum testEnum `long:"enum" required:"true"`
	}{0}

	p := NewNamedParser("test", Default)
	_, err := p.AddCommand("mycmd", "test", "test", &opt)
	if err != nil {
		t.Fatalf("error not expected %+v", err)
	}
	_, err = p.ParseArgs([]string{"mycmd", "--enum=three"})
	if err != nil {
		t.Fatalf("error not expected %+v", err)
	}
	if opt.Enum != three {
		t.Fatalf("expected three, got %v", opt.Enum)
	}

	grp, _ := p.AddGroup("test group", "", &opt)
	o := grp.Options()[0]

	marshalled, err := convertToString(o.value, o.tag)
	if err != nil {
		t.Fatalf("error not expected %+v", err)
	}
	if marshalled != "three" {
		t.Fatalf("expected three, got %v", marshalled)
	}
}
