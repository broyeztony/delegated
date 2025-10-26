package api

import "testing"

func TestValidateYear(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty string - no filter",
			input:   "",
			want:    0,
			wantErr: false,
		},
		{
			name:    "valid year 2022",
			input:   "2022",
			want:    2022,
			wantErr: false,
		},
		{
			name:    "valid year 2000 (minimum)",
			input:   "2000",
			want:    2000,
			wantErr: false,
		},
		{
			name:    "valid year 2100 (maximum)",
			input:   "2100",
			want:    2100,
			wantErr: false,
		},
		{
			name:    "year too low - 1999",
			input:   "1999",
			want:    0,
			wantErr: true,
			errMsg:  "year must be between 2000-2100",
		},
		{
			name:    "year too high - 2101",
			input:   "2101",
			want:    0,
			wantErr: true,
			errMsg:  "year must be between 2000-2100",
		},
		{
			name:    "invalid format - not a number",
			input:   "abc",
			want:    0,
			wantErr: true,
			errMsg:  "year must be a number",
		},
		{
			name:    "partial number - 202",
			input:   "202",
			want:    0,
			wantErr: true,
			errMsg:  "year must be between 2000-2100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateYear(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateYear() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" && err != nil {
				if err.Error() != tt.errMsg {
					t.Errorf("validateYear() error message = %v, want %v", err.Error(), tt.errMsg)
				}
			}

			if got != tt.want {
				t.Errorf("validateYear() = %v, want %v", got, tt.want)
			}
		})
	}
}
