package intel_gpu_top

/*
func Test_jsonFixer(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "empty",
			wantErr: assert.Error,
		},
		{
			name: "single record",
			input: `[
{ "a": 1 }
]
`,
			want: `[
{ "a": 1 }
]
`,
			wantErr: assert.NoError,
		},
		{
			name: "multiple record",
			input: `[
{ "a": 1 }
{ "a": 1 }
{ "a": 1 }
]
`,
			want: `[
{ "a": 1 },
{ "a": 1 },
{ "a": 1 }
]
`,
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := JSONFixer{
				Reader: strings.NewReader(tt.input),
			}
			var got []byte
			for {
				buf := make([]byte, 16)
				n, err := j.Read(buf)
				tt.wantErr(t, err)
				got = append(got, buf[:n]...)
				if n == 0 {
					break
				}
			}
			assert.Equal(t, tt.want, string(got))
		})
	}
}


*/
