package testutil

import "testing"

const succeed = "\u2713"
const failed = "\u2717"

// GSpec test log formatting that's similar to RSpec
type GSpec struct {
	t *testing.T
}

// Given given input
func (gs *GSpec) Given(inputs ...string) {
	gs.t.Helper()
	gs.t.Logf("Given:")
	for _, s := range inputs {
		gs.t.Log("    " + s)
	}
}

// When when describes operation
func (gs *GSpec) When(operation string) {
	gs.t.Helper()
	gs.t.Logf("When:")
	gs.t.Logf("    " + operation)
}

// Then prints out then:
func (gs *GSpec) Then() {
	gs.t.Helper()
	gs.t.Logf("Then:")
}

// Expect calls anonymous code block
func (gs *GSpec) Expect(expect func()) {
	expect()
}

// AssertAndFailNow checks for .Error()assertion and fails now if not true
func (gs *GSpec) AssertAndFailNow(assertion bool, expected string, actual interface{}) {
	gs.t.Helper()
	if assertion != true {
		gs.XMark("%s but got %v", expected, actual)
		gs.t.FailNow()
	}
	gs.CheckMark(expected)
}

// CheckMark prepends checkmark to given message
func (gs *GSpec) CheckMark(format string, args ...interface{}) {
	gs.t.Helper()
	gs.t.Logf("    "+succeed+" "+format, args...)
}

// XMark prepends x mark to given message
func (gs *GSpec) XMark(format string, args ...interface{}) {
	gs.t.Helper()
	gs.t.Logf("    "+failed+" "+format, args...)
}
