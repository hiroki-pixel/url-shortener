package service

import "testing"

func TestIsValidShortCode(t *testing.T) {
	/*
		// Arrange:入力値と期待値を用意する
		testShorCode := "aBC"
		want := false

		// Act:テスト対象の実行
		got := IsValidShortCode(testShorCode)

		//Assert:結果と期待値を比較する
		if got != want {
			t.Errorf("got %t, want %t", got, want)
		}
	*/
	// テーブル駆動テストの実装
	type testCase struct {
		name          string
		testShortCode string
		want          bool
	}
	tests := []testCase{
		{
			name:          "valid_alphanumeric",
			testShortCode: "aB1234",
			want:          true,
		},
		{
			name:          "too_short",
			testShortCode: "abc",
			want:          false,
		},
		{
			name:          "too_long",
			testShortCode: "abc12345",
			want:          false,
		},
		{
			name:          "invalid_charactor",
			testShortCode: "abc@12",
			want:          false,
		},
		{
			name:          "valid_number",
			testShortCode: "123456",
			want:          true,
		},
	}

	// 作成したテストケースから一つずつ入力値と期待値を取り出して、テスト実行する。
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsValidShortCode(tc.testShortCode)

			if got != tc.want {
				t.Errorf(
					"shortCode %q, got %t, want %t",
					tc.testShortCode, got, tc.want,
				)
			}
		})
	}

}
