package aimharder_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xabi93/aimharder/aimharder"
)

const authToken = "141828|1231231|a1b2c3d4e5f6g7"

func TestLogin(t *testing.T) {
	email := "email@email.com"
	password := "super-secure"

	t.Run("wrong endpoint", func(t *testing.T) {
		_, server := muxAndServer(t)

		baseUrl, _ := url.Parse(server.URL + "/")

		_, err := aimharder.Login(context.Background(), email, password,
			aimharder.OptionHttpClient(server.Client()),
			aimharder.OptionBaseURL(baseUrl))

		require.True(t, errors.Is(err, aimharder.EndpointNotExistsError))
	})

	for _, tcases := range []struct {
		testName      string
		apiResp       string
		expectedError error
	}{
		{
			testName:      "wrong username or password",
			apiResp:       `{"error": "Correo electrónico y/o contraseña incorrecto"}`,
			expectedError: aimharder.InvalidMailPassLoginError,
		},
		{
			testName:      "unknown error",
			apiResp:       `{"error": "some_error"}`,
			expectedError: aimharder.UnknownError,
		},
		{
			testName:      "missing auth token",
			apiResp:       `{"cookie": ""}`,
			expectedError: aimharder.MissingAuthTokenError,
		},
	} {
		t.Run(tcases.testName, func(t *testing.T) {
			mux, server := muxAndServer(t)

			mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
				testMethod(t, r, http.MethodGet)

				testQueryParam(t, r, "mail", email)
				testQueryParam(t, r, "pw", password)

				fmt.Fprint(w, tcases.apiResp)
			})

			baseUrl, _ := url.Parse(server.URL + "/")

			_, err := aimharder.Login(context.Background(), email, password,
				aimharder.OptionHttpClient(server.Client()),
				aimharder.OptionBaseURL(baseUrl))

			require.True(t, errors.Is(err, tcases.expectedError), "expected %s error but got %s", tcases.expectedError, err)
		})
	}

	t.Run("success", func(t *testing.T) {
		mux, server := muxAndServer(t)

		mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, http.MethodGet)

			testQueryParam(t, r, "mail", email)
			testQueryParam(t, r, "pw", password)

			fmt.Fprintf(w, `{
				"cookie": %q,
				"userId": "1"
			  }`, authToken)
		})

		baseUrl, _ := url.Parse(server.URL + "/")
		client, err := aimharder.Login(context.Background(), email, password,
			aimharder.OptionHttpClient(server.Client()),
			aimharder.OptionBaseURL(baseUrl))
		require.NoError(t, err)

		require.Equal(t, authToken, client.AuthToken())
	})
}

func setup(t *testing.T) (client *aimharder.Client, mux *http.ServeMux) {
	t.Helper()

	var server *httptest.Server
	mux, server = muxAndServer(t)

	baseUrl, _ := url.Parse(server.URL + "/")

	client, err := aimharder.New(authToken,
		aimharder.OptionHttpClient(server.Client()),
		aimharder.OptionBaseURL(baseUrl),
	)
	if err != nil {
		t.Fatal(err)
	}

	return client, mux
}

func muxAndServer(t *testing.T) (mux *http.ServeMux, server *httptest.Server) {
	t.Helper()

	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	t.Cleanup(func() {
		server.Close()
	})

	return mux, server
}

func testMethod(t *testing.T, r *http.Request, expected string) {
	t.Helper()
	require.Equal(t, expected, r.Method, "expected request to be a %s but was %s", expected, r.Method)
}

func testQueryParam(t *testing.T, r *http.Request, param, expected string) {
	t.Helper()
	v := r.URL.Query().Get(param)
	require.Equal(t, expected, v, "expected request to have %s: %s query param, but had %s", param, expected, v)
}

func testAuthToken(t *testing.T, r *http.Request) {
	t.Helper()
	reqToken := r.URL.Query().Get("token")
	require.Equal(t, authToken, reqToken, "expected request to have %q auth token, but had %q", authToken, reqToken)
}
