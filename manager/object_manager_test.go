package manager

import (
	"testing"
)

type Foo struct {
	Value int
}

func TestObjectManager_Adding_Objects_Sequential(t *testing.T) {
	om := NewObjectManager[Foo]()
	fooObj1Ethalon := Foo{Value: 10}
	id1 := om.AddObject(fooObj1Ethalon)
	got := id1
	if got == om.GetUnknownKey() {
		t.Errorf("AddObject(...) = %d; want not %d", got, om.GetUnknownKey())
	}
	fooObj2Ethalon := Foo{Value: 22}
	id2 := om.AddObject(fooObj2Ethalon)
	got = id2
	if got == om.GetUnknownKey() {
		t.Errorf("AddObject(...) = %d; want not %d", got, om.GetUnknownKey())
	}

	fooObj1Test, error := om.GetObject(id1)
	if error != nil {
		t.Errorf("GetObject(...) return %s; want nil", error.Error())
	}
	if fooObj1Test.Value != fooObj1Ethalon.Value {
		t.Errorf("GetObject(...).Value = %d; want %d", fooObj1Test.Value, fooObj1Ethalon.Value)
	}

	fooObj2Test, error := om.GetObject(id2)
	if error != nil {
		t.Errorf("GetObject(...) return %s; want nil", error.Error())
	}
	if fooObj2Test.Value != fooObj2Ethalon.Value {
		t.Errorf("GetObject(...).Value = %d; want %d", fooObj2Test.Value, fooObj2Ethalon.Value)
	}
}
