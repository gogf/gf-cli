package gen

import (
	"testing"
	"github.com/gogf/gf/test/gtest"
)

func TestSnakeCase(t *testing.T) {
	cases := [][]string{
		{"RGBCodeMd5", "rgb_code_md5"},
		{"testCase", "test_case"},
		{"Md5", "md5"},
		{"userID", "user_id"},
		{"RGB", "rgb"},
		{"RGBCode", "rgb_code"},
		{"_ID", "id"},
		{"User_ID", "user_id"},
		{"user_id", "user_id"},
		{"md5", "md5"},
	}

	for _, i := range cases {
		in := i[0]
		out := i[1]
		result := SnakeToCamelCase(in)
		if result != out {
			gtest.Assert(result, in)
		}
	}
}
