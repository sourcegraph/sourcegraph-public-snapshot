package downloader

import "testing"

func Test_parseCVSS(t *testing.T) {
	tests := []struct {
		name         string
		cvssVector   string
		wantScore    string
		wantSeverity string
		wantErr      bool
	}{
		{
			name:         "Valid CVSS v2.0",
			cvssVector:   "AV:L/AC:M/Au:S/C:P/I:P/A:P",
			wantScore:    "4.1",
			wantSeverity: "MEDIUM",
			wantErr:      false,
		},
		{
			name:         "Valid CVSS v3.0",
			cvssVector:   "CVSS:3.0/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
			wantScore:    "9.8",
			wantSeverity: "CRITICAL",
			wantErr:      false,
		},
		{
			name:         "Valid CVSS v3.1",
			cvssVector:   "CVSS:3.1/AV:A/AC:H/PR:L/UI:R/S:U/C:L/I:L/A:L",
			wantScore:    "4.3",
			wantSeverity: "MEDIUM",
			wantErr:      false,
		},
		{
			name:         "Invalid CVSS v3.1",
			cvssVector:   "CVSS:3.1/AV:A/PR:L/UI:R/S:U/C:L/I:L/A:L",
			wantScore:    "",
			wantSeverity: "",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotScore, gotSeverity, err := parseCVSS(tt.cvssVector)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCVSS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotScore != tt.wantScore {
				t.Errorf("parseCVSS() gotScore = %v, want %v", gotScore, tt.wantScore)
			}
			if gotSeverity != tt.wantSeverity {
				t.Errorf("parseCVSS() gotSeverity = %v, want %v", gotSeverity, tt.wantSeverity)
			}
		})
	}
}
