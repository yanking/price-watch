package log

import "testing"

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Level: "info", Format: "json", Output: "stdout",
			},
			wantErr: false,
		},
		{
			name:    "empty level",
			config:  Config{Level: "", Format: "json", Output: "stdout"},
			wantErr: true,
		},
		{
			name:    "invalid format",
			config:  Config{Level: "info", Format: "xml", Output: "stdout"},
			wantErr: true,
		},
		{
			name:    "empty output",
			config:  Config{Level: "info", Format: "json", Output: ""},
			wantErr: true,
		},
		{
			name:    "invalid output (file path)",
			config:  Config{Level: "info", Format: "json", Output: "/var/log/app.log"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
