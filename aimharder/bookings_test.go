package aimharder_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xabi93/aimharder/aimharder"
)

func TestBookingsService_Next(t *testing.T) {
	boxID := "123"
	t.Run("wrong endpoint", func(t *testing.T) {
		cli, _ := setup(t)
		_, err := cli.Bookings.Next(context.Background(), boxID)

		require.True(t, errors.Is(err, aimharder.EndpointNotExistsError))
	})

	for _, tcases := range []struct {
		testName      string
		apiResp       string
		expectedError error
	}{
		{
			testName:      "unexpected error",
			apiResp:       `{"error": "unknown"}`,
			expectedError: aimharder.UnknownError,
		},
		{
			testName:      "logout",
			apiResp:       `{"logout": 1}`,
			expectedError: aimharder.LogoutError,
		},
	} {
		t.Run(tcases.testName, func(t *testing.T) {
			cli, mux := setup(t)

			mux.HandleFunc("/nextBookings", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)
				testQueryParam(t, r, "box", boxID)
				testAuthToken(t, r)

				fmt.Fprint(w, tcases.apiResp)
			})

			_, err := cli.Bookings.Next(context.Background(), boxID)

			require.True(t, errors.Is(err, tcases.expectedError), "expected %q error but was %q", tcases.expectedError, err)
		})
	}

	t.Run("success", func(t *testing.T) {
		cli, mux := setup(t)

		expectedBookings := []aimharder.Booking{
			{
				ID:        "123",
				Time:      "9:00 - 10:00",
				Day:       "Domingo, 20 de Diciembre de 2020",
				BoxName:   "super-box",
				BookState: "1",
				Waitlist:  "-1",
			},
			{
				ID:        "321",
				Time:      "10:00 - 11:00",
				Day:       "Domingo, 21 de Diciembre de 2020",
				BoxName:   "super-box",
				BookState: "1",
				Waitlist:  "-1",
			},
		}

		mux.HandleFunc("/nextBookings", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			testAuthToken(t, r)

			b, _ := json.Marshal(expectedBookings)
			fmt.Fprintf(w, `{"nextClasses":%s}`, string(b))
		})

		b, err := cli.Bookings.Next(context.Background(), boxID)

		require.NoError(t, err)

		require.Equal(t, expectedBookings, b)
	})
}
