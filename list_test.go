package gotomic

import (
	"reflect"
	"testing"
)

func TestListCreation(t *testing.T) {
	l := NewList()
	checkListEquality(t, l, []thing{})
}

func checkListEquality(t *testing.T, l *List, things []thing) {
	if l.Size() == len(things) {
		i := l.Iterator()
		for _, thing := range things {
			if i.HasNext() {
				v := i.Next()
				if !reflect.DeepEqual(v, thing) {
					t.Error(v, "should equal", thing)
				}
			} else {
				t.Error(l, "should have next")
			}
		}
		if i.HasNext() {
			t.Error(l, "should not have next")
		}
	} else {
		t.Error(l, "should have length", len(things), "but has length", l.Size())
	}
}

func TestListPushing(t *testing.T) {
	l := NewList()
	l.Push("hej")
	checkListEquality(t, l, []thing{"hej"})
	l.Push("på")
	checkListEquality(t, l, []thing{"på", "hej"})
	l.Push("dig")
	checkListEquality(t, l, []thing{"dig", "på", "hej"})
}

func TestListPopping(t *testing.T) {
	l := NewList()
	l.Push("hej")
	l.Push("på")
	l.Push("dig")
	checkListEquality(t, l, []thing{"dig", "på", "hej"})
	if l.Pop() != "dig" {
		t.Error("should be dig")
	}
	checkListEquality(t, l, []thing{"på", "hej"})
	if l.Pop() != "på" {
		t.Error("should be på")
	}
	checkListEquality(t, l, []thing{"hej"})
	if l.Pop() != "hej" {
		t.Error("should be hej")
	}
	checkListEquality(t, l, []thing{})
}
