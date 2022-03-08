package nullable

import "time"

// Bool returns a pointer to a bool value.
func Bool(in bool) *bool { ret := in; return &ret }

// BoolValue returns the value of the bool pointer or false if the pointer is
// nil.
func BoolValue(v *bool) bool {
	if v != nil {
		return *v
	}
	return false
}

// Duration returns a pointer to a duration value.
func Duration(in time.Duration) *time.Duration { ret := in; return &ret }

// Int returns a pointer to an int value.
func Int(in int) *int { ret := in; return &ret }

// Int32 returns a pointer to an int32 value.
func Int32(in int32) *int32 { ret := in; return &ret }

// Int64 returns a pointer to an int64 value.
func Int64(in int64) *int64 { ret := in; return &ret }

// String returns a pointer to a string value.
func String(in string) *string { ret := in; return &ret }

// StringValue returns the value of the string pointer or empty string if the
// pointer is nil.
func StringValue(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}

// StringSlice returns a pointer to a string slice.
func StringSlice(in []string) *[]string { ret := in; return &ret }

// StringStringMap returns a pointer to a string-to-string map.
func StringStringMap(in map[string]string) *map[string]string { ret := in; return &ret }

// Time returns a pointer to a time.Time value.
func Time(in time.Time) *time.Time { ret := in; return &ret }

// TimeValue returns the value of the time pointer or time zero value if the
// pointer is nil.
func TimeValue(in *time.Time) time.Time {
	if in != nil {
		return *in
	}
	return time.Time{}
}

// Uint returns a pointer to a uint value.
func Uint(in uint) *uint { ret := in; return &ret }
