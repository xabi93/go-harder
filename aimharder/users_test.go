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

func TestUsersService_Me(t *testing.T) {
	t.Run("missing user id", func(t *testing.T) {
		cli, _ := aimharder.New("")
		_, err := cli.Users.Me(context.Background())

		require.True(t, errors.Is(err, aimharder.MissingUserIDAuthToken))
	})

	t.Run("wrong endpoint", func(t *testing.T) {
		cli, _ := setup(t)
		_, err := cli.Users.Me(context.Background())

		require.True(t, errors.Is(err, aimharder.EndpointNotExistsError))
	})

	for _, tcases := range []struct {
		testName      string
		apiResp       string
		expectedError error
	}{
		{
			testName:      "user does not exists",
			apiResp:       `{}`,
			expectedError: aimharder.UserNotFound,
		},
		{
			testName:      "logout",
			apiResp:       `{"logout": 1}`,
			expectedError: aimharder.LogoutError,
		},
	} {
		t.Run(tcases.testName, func(t *testing.T) {
			cli, mux := setup(t)

			mux.HandleFunc("/user/141828", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)
				testAuthToken(t, r)

				fmt.Fprint(w, tcases.apiResp)
			})

			_, err := cli.Users.Me(context.Background())

			require.True(t, errors.Is(err, tcases.expectedError), "expected %q error but was %q", tcases.expectedError, err)
		})
	}

	t.Run("success", func(t *testing.T) {
		cli, mux := setup(t)

		expectedUser := aimharder.User{
			ID:          "123",
			Name:        "user-name",
			Alias:       "user-alias",
			Sex:         "F",
			UserFNac:    "11-12-1993",
			Desc:        "user Desc",
			DescNoParse: "user desc no parse",
			Pic:         "http://pic.com/endpoint",
			Box:         "super-box",
			BoxShort:    "super",
			BoxID:       "123",
			Followers:   "11",
			Following:   "22",
			Routines:    "2",
			UserMail:    "user@email.com",
			UserLangUse: "ES",
		}

		mux.HandleFunc("/user/141828", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)
			testAuthToken(t, r)

			b, _ := json.Marshal(expectedUser)
			fmt.Fprint(w, string(b))
		})

		u, err := cli.Users.Me(context.Background())

		require.NoError(t, err)

		require.Equal(t, expectedUser, *u)
	})
}
