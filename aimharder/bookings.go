package aimharder

import (
	"context"
	"net/http"
)

// BookingsService handles communication with the booking methods
type BookingsService service

// Booking represents a Box class reservation
type Booking struct {
	ID        string `json:"id"`
	Time      string `json:"time"`
	ClassName string `json:"className"`
	BoxName   string `json:"boxName"`
	BoxPic    string `json:"boxPic"`
	CoachName string `json:"coachName"`
	BookState string `json:"bookState"`
	Waitlist  string `json:"waitlist"`
	Day       string `json:"day"`
}

//Next list all upcoming class reservations by box
func (s *BookingsService) Next(ctx context.Context, boxID string) ([]Booking, error) {
	req, err := s.client.newRequest(ctx, http.MethodGet, "nextBookings", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("box", boxID)
	req.URL.RawQuery = q.Encode()

	var success struct {
		NextClasses []Booking `json:"nextClasses,omitempty"`
	}
	body := apiBodyDecoder{
		success: &success,
	}

	if _, err := s.client.do(ctx, req, &body); err != nil {
		return nil, err
	}

	if err := body.Error(); err != nil {
		return nil, err
	}

	return success.NextClasses, nil
}
