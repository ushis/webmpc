package webmpc

type Result struct {
  Type string      // The data type
  Data interface{} // The data
}

// Returns a fresh result.
func NewResult(t string, data interface{}) *Result {
  return &Result{t, data}
}
