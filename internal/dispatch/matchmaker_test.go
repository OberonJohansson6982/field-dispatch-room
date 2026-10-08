package dispatch

import (
	"errors"
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name        string
		technicians []Technician
		wantID      string
		wantErr     error
	}{
		{
			name: "lowest active workload wins",
			technicians: []Technician{
				{ID: "tech-b", Online: true, OpenJobs: 3},
				{ID: "tech-a", Online: true, OpenJobs: 1},
				{ID: "tech-c", Online: false, OpenJobs: 0},
			},
			wantID: "tech-a",
		},
		{
			name: "stable ID resolves equal workload",
			technicians: []Technician{
				{ID: "tech-b", Online: true, OpenJobs: 1},
				{ID: "tech-a", Online: true, OpenJobs: 1},
			},
			wantID: "tech-a",
		},
		{
			name: "offline roster cannot take a dispatch",
			technicians: []Technician{
				{ID: "tech-a", Online: false, OpenJobs: 0},
			},
			wantErr: ErrNoTechnician,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Match(tt.technicians)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Match() error = %v, want %v", err, tt.wantErr)
			}
			if got.ID != tt.wantID {
				t.Fatalf("Match() technician = %q, want %q", got.ID, tt.wantID)
			}
		})
	}
}
