package dispatch

import (
	"errors"
	"sort"
)

var ErrNoTechnician = errors.New("no eligible technician is online")

type Technician struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Online      bool   `json:"online"`
	OpenJobs    int    `json:"open_jobs"`
}

// Match selects the online technician with the fewest open jobs. ID breaks ties.
func Match(technicians []Technician) (Technician, error) {
	eligible := make([]Technician, 0, len(technicians))
	for _, technician := range technicians {
		if technician.Online {
			eligible = append(eligible, technician)
		}
	}
	if len(eligible) == 0 {
		return Technician{}, ErrNoTechnician
	}

	sort.Slice(eligible, func(i, j int) bool {
		if eligible[i].OpenJobs == eligible[j].OpenJobs {
			return eligible[i].ID < eligible[j].ID
		}
		return eligible[i].OpenJobs < eligible[j].OpenJobs
	})
	return eligible[0], nil
}
