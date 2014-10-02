package gopool

import "testing"

func TestNewPool(t *testing.T) {
	p := NewPool(10, 5, true)
	expected := 10
	actual := p.PoolSize
	if actual != expected {
		t.Errorf("expected %d, but actually %d", expected, actual)
	}
	expected = 5
	actual = p.QueueCapacity
	if actual != expected {
		t.Errorf("expected %d, but actually %d", expected, actual)
	}
}
